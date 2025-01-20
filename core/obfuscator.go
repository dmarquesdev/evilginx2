package core

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/spf13/viper"
)

type Obfuscator struct {
	Obfuscations []*Obfuscation
	Path         string
}

type Obfuscation struct {
	Name        string
	Description string
	Mimes       []string
	Command     string
	CommandArgs []string
}

func NewObfuscator(path string) (*Obfuscator, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var obs []*Obfuscation

	for _, f := range files {
		if !f.IsDir() {
			or := regexp.MustCompile(`([a-zA-Z0-9\-\.]*)\.yaml`)
			roname := or.FindStringSubmatch(f.Name())
			if roname == nil || len(roname) < 2 {
				continue
			}
			c := viper.New()
			c.SetConfigType("yaml")
			c.SetConfigFile(filepath.Join(path, roname[0]))

			err := c.ReadInConfig()
			if err != nil {
				continue
			}
			obf := &Obfuscation{
				Name:        c.GetString("name"),
				Description: c.GetString("description"),
				Mimes:       c.GetStringSlice("mimes"),
				Command:     c.GetString("command"),
				CommandArgs: c.GetStringSlice("args"),
			}
			obs = append(obs, obf)
		}
	}

	ob := &Obfuscator{
		Obfuscations: obs,
		Path:         path,
	}

	return ob, nil
}

func (o *Obfuscator) obfuscate(data []byte, mime string) []byte {
	obfuscated := data

	for _, ob := range o.Obfuscations {
		if !slices.Contains(ob.Mimes, mime) {
			continue
		}
		args := replacePlaceholders(ob.CommandArgs)

		cmd := exec.Command(ob.Command, args...)
		cmd.Dir = o.Path
		cmd.Stdin = bytes.NewReader(obfuscated)
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		obfuscated = output
	}

	return obfuscated
}

func replacePlaceholders(args []string) []string {
	//TODO apply string substituiton to parameters
	return args
}
