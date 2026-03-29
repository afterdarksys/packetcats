# fuzz_target.star

target = "10.0.0.1"

eth = packet.new_eth(src="11:22:33:44:55:66", dst="aa:bb:cc:dd:ee:ff")
ip = packet.new_ipv4(src="10.0.0.99", dst=target)
tcp = packet.new_tcp(src_port=12345, dst_port=80, flags={"syn": True})
payload = packet.new_payload("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")

# Normally we just assemble, but let's intentionally corrupt the TCP layer
# and payload using a 15% bit-flip intensity to crash the target!
mutated_tcp = fuzz.mutate(tcp, intensity=0.15)
mutated_payload = fuzz.mutate(payload, intensity=0.15)

raw = packet.assemble([eth, ip, mutated_tcp, mutated_payload])
packet.send("en0", raw)

print("Fuzzed packet sent!")
