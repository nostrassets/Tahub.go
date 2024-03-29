package main

// via https://github.com/nbd-wtf/go-nostr
import (
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func main() {
    // Tahub (via Primal) Nostr account
    _, tahubPrimalSk, _ := nip19.Decode("nsec1630frrtduflj7euzef287psu6e8agu6g4dma9j2unkx5nwzdk6nsq8k4k8")
    _, tahubPrimalPk, _ := nip19.Decode("npub1ap207gf8awp87hqnqphufs2kx3rxct8ez65d5r6td65mj4q3pnfs607xxt")
    fmt.Println("tahub pk: ", tahubPrimalPk)
    fmt.Println("tahub sk: ", tahubPrimalSk)

	// Astral Sample Nostr Account
	// _, astralSk, _ := nip19.Decode("nsec1le8lzsxwc6slc7nur72mc7umn628jh23mdzek32axtjtl28shrfqdpzg75")
	// _, astralPk, _ := nip19.Decode("npub1gnm9yexkwxmecsfamzgcqyt6jm2smn3rdz7ta56a2kh8ffsad2ks7t32a7")
	
	// fmt.Println("astral sk: ", astralSk)
	// fmt.Println("astral pk: ", astralPk)

	// * SCRATCH EXAMPLES
	sk := nostr.GeneratePrivateKey()
	pk, _ := nostr.GetPublicKey(sk)
	// print raw
	fmt.Println("sk:", sk)
	fmt.Println("pk:", pk)
	nsec, _ := nip19.EncodePrivateKey(sk)
	npub, _ := nip19.EncodePublicKey(pk)
	// print encoded
	fmt.Println(nsec)
	fmt.Println(npub)
}



// TODO reset PRIMAL NSEC
//_, skHex, _ := nip19.Decode("nsec1ulhapvxtf6lhqq8ztkm3zz0tmsu052ersz3rm7m0399tss08t7js96qq30")
// fmt.Println("sk:", sk)
// fmt.Println("pk:", pk)
// fmt.Println(nsec)
// fmt.Println(npub)

// TODO reset first PRIMAL
//fmt.Println(skHex)

