package api

import (
	"fmt"
	"strings"

	"github.com/kubernetes-csi/csi-proxy/v2/pkg/cim"
	"github.com/kubernetes-csi/csi-proxy/v2/pkg/utils"
	"github.com/microsoft/wmi/pkg/base/query"
)

type HostAPI interface {
	IsSMBMapped(remotePath string) (bool, error)
	NewSMBLink(remotePath, localPath string) error
	NewSMBGlobalMapping(remotePath, username, password string) error
	RemoveSMBGlobalMapping(remotePath string) error
}

type smbAPI struct{}

var _ HostAPI = &smbAPI{}

func New() HostAPI {
	return smbAPI{}
}

func remotePathForQuery(remotePath string) string {
	return strings.ReplaceAll(remotePath, "\\", "\\\\")
}

func (smbAPI) IsSMBMapped(remotePath string) (bool, error) {
	smbQuery := query.NewWmiQuery("MSFT_SmbGlobalMapping", "RemotePath", remotePathForQuery(remotePath))
	instances, err := cim.QueryInstances(cim.WMINamespaceSmb, smbQuery)
	if cim.IgnoreNotFound(err) != nil {
		return false, err
	}

	return len(instances) > 0, nil
}

// NewSMBLink - creates a directory symbolic link to the remote share.
// The os.Symlink was having issue for cases where the destination was an SMB share - the container
// runtime would complain stating "Access Denied". Because of this, we had to perform
// this operation with powershell commandlet creating an directory softlink.
// Since os.Symlink is currently being used in working code paths, no attempt is made in
// alpha to merge the paths.
// TODO (for beta release): Merge the link paths - os.Symlink and Powershell link path.
func (smbAPI) NewSMBLink(remotePath, localPath string) error {
	if !strings.HasSuffix(remotePath, "\\") {
		// Golang has issues resolving paths mapped to file shares if they do not end in a trailing \
		// so add one if needed.
		remotePath = remotePath + "\\"
	}

	cmdLine := `New-Item -ItemType SymbolicLink $Env:smblocalPath -Target $Env:smbremotepath`
	output, err := utils.RunPowershellCmd(cmdLine, fmt.Sprintf("smbremotepath=%s", remotePath), fmt.Sprintf("smblocalpath=%s", localPath))
	if err != nil {
		return fmt.Errorf("error linking %s to %s. output: %s, err: %v", remotePath, localPath, string(output), err)
	}

	return nil
}

func (smbAPI) NewSMBGlobalMapping(remotePath, username, password string) error {
	// use PowerShell Environment Variables to store user input string to prevent command line injection
	// https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_environment_variables?view=powershell-5.1
	cmdLine := fmt.Sprintf(`$PWord = ConvertTo-SecureString -String $Env:smbpassword -AsPlainText -Force` +
		`;$Credential = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $Env:smbuser, $PWord` +
		`;New-SmbGlobalMapping -RemotePath $Env:smbremotepath -Credential $Credential -RequirePrivacy $true`)

	if output, err := utils.RunPowershellCmd(cmdLine, fmt.Sprintf("smbuser=%s", username),
		fmt.Sprintf("smbpassword=%s", password),
		fmt.Sprintf("smbremotepath=%s", remotePath)); err != nil {
		return fmt.Errorf("NewSMBGlobalMapping failed. output: %q, err: %v", string(output), err)
	}
	return nil
	//params := map[string]interface{}{
	//	"RemotePath":     remotePath,
	//	"RequirePrivacy": api.RequirePrivacy,
	//}
	//if username != "" {
	//	params["Credential"] = fmt.Sprintf("%s:%s", username, password)
	//}
	//result, _, err := cim.InvokeCimMethod(cim.WMINamespaceSmb, "MSFT_SmbGlobalMapping", "Create", params)
	//if err != nil {
	//	return fmt.Errorf("NewSmbGlobalMapping failed. result: %d, err: %v", result, err)
	//}
	//return nil
}

func (smbAPI) RemoveSMBGlobalMapping(remotePath string) error {
	smbQuery := query.NewWmiQuery("MSFT_SmbGlobalMapping", "RemotePath", remotePathForQuery(remotePath))
	instances, err := cim.QueryInstances(cim.WMINamespaceSmb, smbQuery)
	if err != nil {
		return err
	}

	_, err = instances[0].InvokeMethod("Remove", true)
	if err != nil {
		return fmt.Errorf("error remove smb mapping '%s'. err: %v", remotePath, err)
	}

	return nil
}
