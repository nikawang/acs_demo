import base64
import hashlib
import hmac
import json
from datetime import datetime, timezone
from urllib import request


def compute_content_hash(content):
    sha_256 = hashlib.sha256()
    sha_256.update(content)
    hashed_bytes = sha_256.digest()
    base64_encoded_bytes = base64.b64encode(hashed_bytes)
    content_hash = base64_encoded_bytes.decode('utf-8')
    return content_hash

def compute_signature(string_to_sign, secret):
    decoded_secret = base64.b64decode(secret)
    encoded_string_to_sign = string_to_sign.encode('ascii')
    hashed_bytes = hmac.digest(decoded_secret, encoded_string_to_sign, digest=hashlib.sha256)
    encoded_signature = base64.b64encode(hashed_bytes)
    signature = encoded_signature.decode('utf-8')
    return signature

def format_date(dt):
    days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
    months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    utc = dt.utctimetuple()

    return "{}, {:02} {} {:04} {:02}:{:02}:{:02} GMT".format(
    days[utc.tm_wday],
    utc.tm_mday,
    months[utc.tm_mon-1],
    utc.tm_year,
    utc.tm_hour, 
    utc.tm_min, 
    utc.tm_sec)


host = "acstestingmails.communication.azure.com"
resource_endpoint = "https://"+host
path_and_query = "/emails:send?api-version=2021-10-01-preview"
secret = "LN7TgCHRTuM1qCl/uSqFilQ8yCxiQ1agZrFkUJtPSESdbpcxxf9Ejlzg/wUcRGQi3sql6xx9cIvsF3TOuiecDg=="

# Create a uri you are going to call.
request_uri = resource_endpoint+path_and_query
print(request_uri+"\n")

# Endpoint identities?api-version=2021-03-07 accepts list of scopes as a body.
body = {
  "headers": [
    {
      "name": "ClientCorrelationId",
      "value": "123"
    },
    {
      "name": "ClientCustomHeaderName",
      "value": "ClientCustomHeaderValue"
    }
  ],
  "content": {
    "subject": "An exciting offer especially for you!",
    "plainText": "This exciting offer was created especially for you, our most loyal customer.",
    "html": "<html><head><title>Exciting offer!</title></head><body><h1>This exciting offer was created especially for you, our most loyal customer.</h1></body></html>"
  },
  "importance": "normal",
  "recipients": {
    "to": [
      {
        "email": "danielwzhg@gmail.com",
        "displayName": "Daniel Wang"
      }
    ],
    "CC": [
      {
        "email": "zhaw@microsoft.com",
        "displayName": "Zhanggui Wang"
      }
    ]
  }
}

serialized_body = json.dumps(body)
content = serialized_body.encode("utf-8")
print(serialized_body+"\n")
# Specify the 'x-ms-date' header as the current UTC timestamp according to the RFC1123 standard
utc_now = datetime.now(timezone.utc)
date = format_date(utc_now)
# print(date+"\n")
# Compute a content hash for the 'x-ms-content-sha256' header.
content_hash = compute_content_hash(content)
# print(content_hash+"\n")

# Prepare a string to sign.
string_to_sign = "POST\n"+path_and_query+"\n"+date+";"+host+";"+content_hash
print(string_to_sign+"\n")
# Compute the signature.
signature = compute_signature(string_to_sign, secret)
# Concatenate the string, which will be used in the authorization header.
authorization_header = "HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature="+signature

print(authorization_header)


request_headers = {}

# Add a date header.
request_headers["x-ms-date"] = date

# Add content hash header.
request_headers["x-ms-content-sha256"] = content_hash

# Add authorization header.
request_headers["Authorization"] = authorization_header

# Add content type header.
request_headers["Content-Type"] = "application/json"


req = request.Request(request_uri, content, request_headers, method='POST')
with request.urlopen(req) as response:
  response_string = json.load(response)
print(response_string)