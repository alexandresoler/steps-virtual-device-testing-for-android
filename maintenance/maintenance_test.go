package maintenance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/pkg/errors"
)

func TestDeviceList(t *testing.T) {
	signedIn, err := checkAccounts()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !signedIn {
		if err := signIn(); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	if err := checkDeviceList(); err != nil {
		t.Error(err)
	}
}

func checkDeviceList() error {
	cmd := command.New("gcloud", "firebase", "test", "android", "models", "list", "--format", "text", "--filter=VIRTUAL")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return errors.Wrap(err, out)
	}

	if out == deviceList {
		return nil
	}

	cmd = command.New("gcloud", "firebase", "test", "android", "models", "list", "--filter=VIRTUAL")
	outFormatted, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return errors.Wrap(err, out)
	}

	fmt.Println("Fresh devices list to use in this integration test:")
	fmt.Println(out)
	fmt.Println()
	fmt.Println("Fresh device list to use in the step's descriptor:")
	fmt.Println(outFormatted)

	return fmt.Errorf("device list has changed, update the corresponding step descriptor blocks")
}

func signIn() error {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("_serv_acc_")
	if err != nil {
		return err
	}

	servAccFileContent := os.Getenv("SERVICE_ACCOUNT_JSON")
	if servAccFileContent == "" {
		return fmt.Errorf("$SERVICE_ACCOUNT_JSON is not set")
	}

	servAccFilePAth := filepath.Join(tmpDir, "serv-acc.json")
	if err := fileutil.WriteStringToFile(servAccFilePAth, servAccFileContent); err != nil {
		return err
	}

	var servAcc struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.NewDecoder(strings.NewReader(servAccFileContent)).Decode(&servAcc); err != nil {
		return err
	}
	if servAcc.ProjectID == "" {
		return fmt.Errorf("invalid service account json, no project_id found")
	}

	cmd := command.New("gcloud",
		"auth",
		"activate-service-account",
		fmt.Sprintf("--key-file=%s", servAccFilePAth),
		"--project", servAcc.ProjectID)

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	return errors.Wrap(err, out)
}

func checkAccounts() (bool, error) {
	cmd := command.New("gcloud", "auth", "list", "--format", "json")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return false, err
	}

	var accounts []interface{}
	if err := json.NewDecoder(strings.NewReader(out)).Decode(&accounts); err != nil {
		return false, err
	}

	return len(accounts) > 0, nil
}

const deviceList = `---
brand:                  google
codename:               Pixel8
form:                   PHYSICAL
formFactor:             PHONE
id:                     Pixel8
manufacturer:           Google
name:                   Pixel 8
screenDensity:          428
screenX:                1080
screenY:                2400
supportedVersionIds[0]: 35`
