package vasinstruments

import (
	"fmt"
	"testing"
	"time"
)

func (pt *PTraktype) TestGetPTrakdata(t *testing.T) {
	var d int32
	for i := 0; i < 100; i++ {
		for pt.mdata < 0 {
			time.Sleep(time.Second / 10)
			d = pt.GetPTrakdata()
		}
		fmt.Println(d)
	}
}
