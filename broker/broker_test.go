package broker

import (
	"github.com/mission-liao/dingo/common"
	"github.com/mission-liao/dingo/transport"
	"github.com/stretchr/testify/suite"
)

type BrokerTestSuite struct {
	suite.Suite

	_invoker       transport.Invoker
	_producer      Producer
	_consumer      Consumer
	_namedConsumer NamedConsumer
}

func (me *BrokerTestSuite) SetupSuite() {
	me._invoker = transport.NewDefaultInvoker()
	me.NotNil(me._invoker)
}

func (me *BrokerTestSuite) TearDownSuite() {
	me.Nil(me._producer.(common.Object).Close())
}

func (me *BrokerTestSuite) AddListener(name string, receipts <-chan *Receipt) (tasks <-chan []byte, err error) {
	if me._consumer != nil {
		tasks, err = me._consumer.AddListener(receipts)
	} else if me._namedConsumer != nil {
		tasks, err = me._namedConsumer.AddListener("", receipts)
	}

	return
}

func (me *BrokerTestSuite) StopAllListeners() (err error) {
	if me._consumer != nil {
		err = me._consumer.StopAllListeners()
	} else if me._namedConsumer != nil {
		err = me._namedConsumer.StopAllListeners()
	}

	return
}

//
// test cases
//

func (me *BrokerTestSuite) TestBasic() {
	_ = "breakpoint"
	var (
		tasks <-chan []byte
	)
	// init one listener
	receipts := make(chan *Receipt, 10)
	tasks, err := me.AddListener("", receipts)
	me.Nil(err)
	me.NotNil(tasks)
	if tasks == nil {
		return
	}

	// compose a task
	t, err := me._invoker.ComposeTask("", []interface{}{})
	me.Nil(err)
	me.NotNil(t)
	if t == nil {
		return
	}

	// send it
	input := append(
		transport.EncodeHeader(t.ID(), t.Name(), transport.Encode.Default),
		[]byte("test byte array")...,
	)
	me.Nil(me._producer.Send(t, input))

	// receive it
	output := <-tasks
	me.Equal(string(input), string(output))

	// send a receipt
	receipts <- &Receipt{
		ID:     t.ID(),
		Status: Status.OK,
	}

	// stop all listener
	me.Nil(me.StopAllListeners())
}