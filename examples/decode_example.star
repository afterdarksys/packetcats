# decode_example.star - Decodes STDIN packets

def packet_hook(raw):
    print("Received packet of length " + str(len(raw)))
    # You can access base64, json, etc.
    # encoded = base64.encode(raw)
    # print(encoded)
