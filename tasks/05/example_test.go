package solution

import "testing"

func TestStatusCheckOfNonExistingOrder(t *testing.T) {
	client := NewClient("panda")
	if client.CheckStatus("panda") != DOES_NOT_EXIST {
		t.Error("Status check of non existant order failed")
	}
}

func TestOrderIsEnqueued(t *testing.T) {
	factory := NewFactory(0)
	client := NewClient("panda")
	orderId := client.Order(factory, []string{"foo", "bar"})
	factory.StopProducing()

	if _, ok := client.Orders[orderId]; !ok {
		t.Errorf("Order %s is not in the client's orders", orderId)
	}

	if client.CheckStatus(orderId) != ENQUEUED {
		t.Errorf("Order %s is not in the factory's queue", orderId)
	}
}

func TestOrderWithSeveralWords(t *testing.T) {
	factory := NewFactory(0)
	factory.StopProducing()
	factory.StorageAdd(map[string]uint16{
		`che`:             1,
		`(kop|bob|top)`:   1,
		`(panda|raccoon,`: 1,
		`[0-9]+`:          1,
	})
	result, err := factory.generateRegexp([]string{"kopche", "bobche", "topche"})

	if err != nil {
		t.Error("Unable to produce with valid materials")
	}

	if result != `^(kop|bob|top)che$` {
		t.Errorf("Expected ^(kop|bob|top)che$, got %s", result)
	}
}
