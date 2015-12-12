package dingo

import (
	"time"

	"github.com/mission-liao/dingo/common"
	"github.com/mission-liao/dingo/transport"
	"github.com/stretchr/testify/suite"
)

type DingoTestSuite struct {
	suite.Suite

	cfg    *Config
	app    *App
	eid    int
	events <-chan *common.Event
}

func (me *DingoTestSuite) SetupSuite() {
	var err error
	me.cfg = Default()
	me.app, err = NewApp("")
	me.Nil(err)
	me.eid, me.events, err = me.app.Listen(common.InstT.ALL, common.ErrLvl.DEBUG, 0)
	me.Nil(err)
}

func (me *DingoTestSuite) TearDownTest() {
	for {
		done := false
		select {
		case v, ok := <-me.events:
			if !ok {
				done = true
				break
			}
			me.T().Errorf("receiving event:%+v", v)

		// TODO: how to know that there is some error sent...
		// not a good way to block until errors reach here.
		case <-time.After(1 * time.Second):
			done = true
		}

		if done {
			break
		}
	}
}

func (me *DingoTestSuite) TearDownSuite() {
	me.app.StopListen(me.eid)
	me.Nil(me.app.Close())
}

//
// test cases
//

func (me *DingoTestSuite) TestBasic() {
	// register a set of workers
	_ = "breakpoint"
	called := 0
	err := me.app.Register("TestBasic",
		func(n int) int {
			called = n
			return n + 1
		}, transport.Encode.Default, transport.Encode.Default,
	)
	me.Nil(err)
	remain, err := me.app.Allocate("TestBasic", 1, 1)
	me.Nil(err)
	me.Equal(0, remain)

	// call that function
	reports, err := me.app.Call("TestBasic", nil, 5)
	me.Nil(err)
	me.NotNil(reports)

	// await for reports
	status := []int16{
		transport.Status.Sent,
		transport.Status.Progress,
		transport.Status.Done,
	}
	for {
		done := false
		select {
		case v, ok := <-reports:
			me.True(ok)
			if !ok {
				break
			}

			// make sure the order of status is right
			me.True(len(status) > 0)
			if len(status) > 0 {
				me.Equal(status[0], v.Status())
				status = status[1:]
			}

			if v.Done() {
				me.Equal(5, called)
				me.Len(v.Return(), 1)
				if len(v.Return()) > 0 {
					ret, ok := v.Return()[0].(int)
					me.True(ok)
					me.Equal(called+1, ret)
				}
				done = true
			}
		}

		if done {
			break
		}
	}
}
