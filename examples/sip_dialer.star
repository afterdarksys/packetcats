# sip_dialer.star

target = "10.0.0.10"
target_port = "5060"
call_id = "12345-abcde-67890"

# Generate SIP INVITE
sdp = "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=session\r\nc=IN IP4 127.0.0.1\r\nt=0 0\r\nm=audio 8000 RTP/AVP 0\r\na=rtpmap:0 PCMU/8000\r\n"
invite_payload = sip.invite(
    request_uri="sip:1000@" + target,
    to="<sip:1000@" + target + ">",
    from="<sip:attacker@evil.com>",
    call_id=call_id,
    sdp=sdp,
    local_ip="127.0.0.1",
    local_port="5060"
)

print("Generated SIP INVITE:\n")
print(invite_payload)

# If we wanted to fire this off via raw injection:
# eth = packet.new_eth(...)
# ip = packet.new_ipv4(dst=target)
# udp = packet.new_udp(dst_port=5060)
# payload = packet.new_payload(invite_payload)
# packet.send("en0", packet.assemble([eth, ip, udp, payload]))

# Once the call establishes, we could stream RTP audio directly over standard UDP sockets!
# (Assuming the target negotiated port 8000 from the PBX)
# print("Streaming RTP Audio...")
# rtp.stream_wav("rickroll.wav", target_ip=target, target_port=8000)
