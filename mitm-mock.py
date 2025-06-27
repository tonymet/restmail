from mitmproxy import http

class BlockUnknown:
    def request(self, flow: http.HTTPFlow) -> None:
        # If this request is not being replayed, kill it
        if not flow.is_replay:
            # Option 1: Kill the request (client gets connection closed)
            #flow.kill()
            # Option 2: Or return a custom error response instead of killing
            flow.response = http.Response.make(
                404, b"Not found in replay", {"Content-Type": "text/plain"}
            )

addons = [BlockUnknown()]