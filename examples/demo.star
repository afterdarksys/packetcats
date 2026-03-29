# demo.star - A simple SYN generator

target_ip = net.resolve("example.com")
my_ip = net.my_ip()

print("Target: " + target_ip + ", Source: " + my_ip)

eth = packet.new_eth(src="00:11:22:33:44:55", dst="aa:bb:cc:dd:ee:ff")
ip_layer = packet.new_ipv4(src=my_ip, dst=target_ip)
tcp_layer = packet.new_tcp(src_port=54321, dst_port=80, flags={"syn": True})

raw = packet.assemble([eth, ip_layer, tcp_layer])
print("Built raw packet of size: " + str(len(raw)))
