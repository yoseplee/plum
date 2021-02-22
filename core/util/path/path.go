package path

import (
	"fmt"
	"os"
	"sync"
)

type Path struct {
	PlumRoot         string
	ProfilePath      string
	GenesisBlockPath string
	LedgerPath       string
}

var (
	once sync.Once
	path *Path
)

func GetInstance() *Path {
	once.Do(func() {
		path = new(Path)
	})
	return path
}

func init() {
	p := GetInstance()
	p.PlumRoot = os.ExpandEnv("$PLUM_ROOT")
	p.ProfilePath = fmt.Sprintf("%s%s", p.PlumRoot, "/core/profile.yaml")
	p.GenesisBlockPath = fmt.Sprintf("%s%s", p.PlumRoot, "/core/genesis.block")
	p.LedgerPath = fmt.Sprintf("%s%s", p.PlumRoot, "/ledger_store/")
}
