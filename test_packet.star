target_ip = net.resolve("example.com")
my_ip = net.my_ip()

print("Crafting from " + my_ip + " to " + target_ip)

eth = packet.new_eth(src="00:11:22:33:44:55", dst="aa:bb:cc:dd:ee:ff")
ip = packet.new_ipv4(src=my_ip, dst=target_ip)
tcp = packet.new_tcp(src_port=12345, dst_port=80, flags={"syn": True})
payload = packet.new_payload(data="GET / HTTP/1.0\r\n\r\n")

raw = packet.assemble([eth, ip, tcp, payload])
print("Assembled packet of size: " + str(len(raw)))
