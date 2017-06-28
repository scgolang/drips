package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/scgolang/sc"
	"github.com/scgolang/scid"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	client, err := sc.NewClient("udp", "127.0.0.1:0", "127.0.0.1:57120", 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	if err := drips(client); err != nil {
		log.Fatal(err)
	}
}

func drips(client *sc.Client) error {
	// Load reverb synthdef.
	if err := client.SendDef(sc.DefJPverbMono); err != nil {
		return err
	}
	// Load drips synthdef.
	if err := client.SendDef(def); err != nil {
		return err
	}

	// Create effect group.
	effectGroup, err := client.AddDefaultGroup()
	if err != nil {
		return err
	}

	// Create synth group.
	// synthGroup, err := client.Group(int32(2), sc.AddToHead, sc.DefaultGroupID)
	synthGroup, err := client.Group(int32(2), sc.AddToHead, sc.DefaultGroupID)
	if err != nil {
		log.Fatal(err)
	}

	// Create the reverb node.
	var (
		bus      = float32(16)
		verbCtls = map[string]float32{"in": bus}
	)
	sid, err := scid.Next()
	if err != nil {
		return err
	}
	if _, err := effectGroup.Synth(sc.DefJPverbMono.Name, sid, sc.AddToTail, verbCtls); err != nil {
		return err
	}

	// Create the drips node.
	for {
		dripCtls := map[string]float32{
			"freq": float32((200 * rand.Float64()) + 400),
			"out":  bus,
		}
		sid, err = scid.Next()
		if err != nil {
			return err
		}
		if _, err := synthGroup.Synth(def.Name, sid, sc.AddToTail, dripCtls); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func simple(client *sc.Client) error {
	g, err := client.AddDefaultGroup()
	if err != nil {
		return err
	}
	if err := client.SendDef(sc.NewSynthdef("drips", func(params sc.Params) sc.Ugen {
		return sc.Out{
			Bus: sc.C(0),
			Channels: sc.JPverb{
				In: sc.Saw{}.Rate(sc.AR),
			}.Rate(sc.AR),
		}.Rate(sc.AR)
	})); err != nil {
		return err
	}
	sid, err := scid.Next()
	if err != nil {
		return err
	}
	_, err = g.Synth(def.Name, sid, sc.AddToTail, nil)
	return err
}

var def = sc.NewSynthdef("drips", func(params sc.Params) sc.Ugen {
	var (
		freq = params.Add("freq", 440)
		out  = params.Add("out", 0)
	)
	env := sc.EnvGen{
		Env:  sc.EnvPerc{},
		Done: sc.FreeEnclosing,
	}.Rate(sc.KR)

	sin := sc.SinOsc{Freq: freq}.Rate(sc.AR)

	return sc.Out{
		Bus:      out,
		Channels: sin.Mul(env),
	}.Rate(sc.AR)
})
