# honeypot.star

target_port = 2222

print("Starting Fake SSH Honeypot on Port 2222...")
print("PacketCats TCP Stack is now automatically responding to SYNs with SYN-ACKs!")

# This drops a background BPF listener. It will catch inbound SYNs to 2222 
# and instantly generate a valid SYN-ACK packet so the remote scanner thinks
# we have an open SSH port!
tcpstack.listen_syn_ack("en0", target_port)

# We can then wait or run other logic, while the engine tarpits scanners
while True:
    # Await in a real script...
    pass
