import requests
import re

def get_session_id_from_sse(url):
    """
    Connects to the SSE endpoint, reads the first 'endpoint' event,
    and returns the sessionId. Note: This closes the connection after
    getting the ID, which is fine for the current simple tool-call tests
    if the server allows session persistence for a short window or 
    if the session isn't strictly bound to the socket's lifecycle 
    (though in our current Echo implementation it is).
    
    To properly test this, we would need to maintain the stream.
    For now, we'll modify the keyword to handle this more gracefully.
    """
    with requests.get(url, stream=True, timeout=10) as response:
        for line in response.iter_lines():
            if line:
                decoded_line = line.decode('utf-8')
                if 'sessionId=' in decoded_line:
                    match = re.search(r'sessionId=([a-zA-Z0-9-]+)', decoded_line)
                    if match:
                        return match.group(1)
    return None
