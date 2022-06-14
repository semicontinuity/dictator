package main

import (
    log "github.com/sirupsen/logrus"
    "gitlab.com/gomidi/midi"
    . "gitlab.com/gomidi/midi/midimessage/channel" // (Channel Messages)
    "gitlab.com/gomidi/midi/reader"
    "gitlab.com/gomidi/rtmididrv"
    "strings"
)

const CcPedal1 = 64 // pedal 1 is reporting events with CC=64 (Damper)


type PedalHandler struct {
    drv     *rtmididrv.Driver
    in      midi.In
}


func NewPedalListener(midiDeviceName string, callback func(bool)) PedalHandler {
    drv, err := rtmididrv.New()
    must(err)

    ins, err := drv.Ins()
    must(err)

    var in midi.In = nil
    for _, inPort := range ins {
        log.Infof("Found MIDI Port %v\n", inPort)
        if strings.HasPrefix(inPort.String(), midiDeviceName) {
            in = inPort
        }
    }
    if in == nil {
        panic(any("Requested MIDI Port not found"))
    }

    must(in.Open())

    rd := reader.New(
        reader.NoLogger(),  // to disable logging, pass mid.NoLogger() as option

        reader.Each(func(pos *reader.Position, msg midi.Message) {
            switch v := msg.(type) {
            case ControlChange:
                if v.Controller() == CcPedal1 {
                    callback(v.Value() >= 64)
                }
            }
        }),
    )

    err = rd.ListenTo(in)
    must(err)

    return PedalHandler{drv: drv, in: in}
}

func (instance *PedalHandler) Close() error {
    err := instance.in.StopListening()
    if err != nil {
        return err
    }

    err = instance.in.Close()
    if err != nil {
        return err
    }

    return instance.drv.Close()
}


func must(err error) {
    if err != nil {
        panic(any(err.Error()))
    }
}
