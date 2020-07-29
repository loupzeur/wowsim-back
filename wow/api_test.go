package wow

import (
	"fmt"
	"testing"
)

func Test_auth(t *testing.T) {
	//auth()
}

func Test_getCharacterEquipment(t *testing.T) {
	z := GetCharacterEquipment("eu", "arathi", "grosmatt")
	fmt.Printf("%+v", z)
}
