package scripting

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// PCAPModule returns the "pcap" Starlark module
func PCAPModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "pcap",
		Members: starlark.StringDict{
			"replay": starlark.NewBuiltin("replay", pcapReplay),
		},
	}
}

func pcapReplay(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var filename, iface string
	var speed float64 = 1.0
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "filename", &filename, "interface", &iface, "speed?", &speed); err != nil {
		return nil, err
	}

	if speed <= 0 {
		return nil, fmt.Errorf("speed must be greater than 0")
	}

	outHandle, err := pcap.OpenLive(iface, 65536, false, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("error opening output interface %s: %v", iface, err)
	}
	defer outHandle.Close()

	inHandle, err := pcap.OpenOffline(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening pcap file %s: %v", filename, err)
	}
	defer inHandle.Close()

	packetSource := gopacket.NewPacketSource(inHandle, inHandle.LinkType())
	
	var lastTime time.Time
	
	count := 0
	for packet := range packetSource.Packets() {
		currentTime := packet.Metadata().Timestamp
		
		if !lastTime.IsZero() && speed < 1000.0 { // speed >= 1000 means basically as fast as possible
			delay := currentTime.Sub(lastTime)
			if delay > 0 {
				sleepTime := time.Duration(float64(delay) / speed)
				time.Sleep(sleepTime)
			}
		}

		if err := outHandle.WritePacketData(packet.Data()); err != nil {
			return nil, fmt.Errorf("failed to write packet to %s: %v", iface, err)
		}

		lastTime = currentTime
		count++
	}

	return starlark.MakeInt(count), nil
}
