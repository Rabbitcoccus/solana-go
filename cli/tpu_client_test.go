package cli

import (
	"testing"

	"github.com/Rabbitcoccus/solana-go"
	"github.com/Rabbitcoccus/solana-go/rpc"
	"github.com/Rabbitcoccus/solana-go/rpc/tpu"
)

func TestClient(t *testing.T) {
	rpcCli, err := rpc.New("https://fabled-damp-thunder.solana-mainnet.quiknode.pro/e3a2acc1e0273bf2d35fefbd1170bda5f00df04c/")
	if err != nil {
		t.Error("failed to create RPC client: ", err)
		return
	}
	tpuConfig := tpu.TPUClientConfig{
		FanoutSlots: tpu.DEFAULT_FANOUT_SLOTS,
	}
	tpuCli, err := tpu.New(rpcCli, "wss://fabled-damp-thunder.solana-mainnet.quiknode.pro/e3a2acc1e0273bf2d35fefbd1170bda5f00df04c/", tpuConfig)
	if err != nil {
		t.Error("failed to create TPU client: ", err)
		return
	}
	tx := new(solana.Transaction)
	err = tx.UnmarshalBase64(txb64)
	if err != nil {
		t.Error(err)
		return
	}
	tpuCli.SendTransaction(tx, 1)
}

const txb64 = `ATI5nPZESFTzOzZjEcXn4+TdGwsQSmVUz1WI+0rweMUlN+loXDTbeMgBXtY/GcQu/kbhF3L1EA8rhGGbbh3q0g6AAQAJFKApBKNIB4MgIF3PsuWiVaoHMaSRUxIDkLdjVkLvrPcACgUkQaFYTleMcAX2p74eBXQZd1dwDyQZAPJfSv2KGc4ORVl2xu2CjBEPfICWXBP/+1cUhXXATjIzgQH+R97wVyGGGsDJpY3/WI6D+HzzFAT2jryNjRgU9DQt1WagiOVETeXmPFT6bTd9o6hhn/Ng4zdCCV6HpuHmzVQKYKjuwrJdm0WD4kesF2RaG4DOkjwLxdwkitJUTYl7f2pyXXap8YyiqmhuwMc7RmmUVvoMuaH3iBqXGQ5G3sZVPE1FKZl7oiPWBx0Ge9srKsZl98a8+rXd7D2hyWzl5I4CIeWDSPS5ha48uggePWu090s1ff4yoCjAvULeZ+cqYFn+Adk5TejkCTOoGbfqUBt54ouIr4cw03h1ecqVqS4W5yx4v9R9+PTko52VL0CIM2xtl0WkvNslD6Wawxr7yd9HYllN4LwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMGRm/lIRcy/+ytunLDm+e8jOW7xfcSayxDmzpAAAAABHnVW/IxwG7udMVuzmgVB/2xst6j9I5RArHNola8E48Gm4hX/quBhPtof2NGGMA12sQ53BrrO1WYoPAAAAAAAQbd9uHXZaGT2cvhRs7reawctIXtX1s3kTqM9YV+/wCpjJclj04kifG7PRApFI4NgwtaE5na/xCEBI572Nvp+Fm0P/on9df2SnTAmx8pWHneSwmrNt/J3VFLMhqns4zl6M4BDmCv7bInF71jGS9UFFo/llozu4LSxwKess4eIIJk4vh36D2uxoiPQ6u1rnsxG7eMNJYIl/zlp7gsFuIP8ZeHmQH9/FCUkE69T0VdnJHYw4UQGk1bQLzCav1HaZBoogUMAAUCwFwVAAwACQNbzBUAAAAAABAGAAcADgsPAQENNQ8TAAMKAQcSDgQNEQ0kJQ8mEwoIHB4dIg8XIRsYGiMZFxcXFxcXCAYTIA8TFgEUBhUCBQkfLcEgmzNB1pyBAAMAAAAKZAABB2QBAhEAZAIDeV4AAAAAAACK+gEAAAAAAIgTMg8DBwAAAQkDbe3JDtbl0cv3RLd1uFdnrlOG+ym60F615QMRqGftcwkDRkQ/BDxFDATmRHYdvDa2+/HneaNZWFJUkYpSl+ontm5e0vRcHOnqDQU87vDt8QE32aCCxZO4HxNsuW7dwAvoP+DsDc3YhygYSG3XmL2BRI0DysjJA8TFxw==`
