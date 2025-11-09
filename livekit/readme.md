Your production config files are generated in directory: video.recontext.online

Please point update DNS for the following domains to the IP address of your server.
* video.recontext.online
* turn.recontext.online
* ingress.recontext.online
  Once started, Caddy will automatically acquire TLS certificates for the domains.

The file "cloud_init.ubuntu.yaml" is a script that can be used in the "user-data" field when starting a new VM.

Since you've enabled Egress/Ingress, we recommend running it on a machine with at least 4 cores

Please ensure the following ports are accessible on the server
* 443 - primary HTTPS and TURN/TLS
* 80 - for TLS issuance
* 7881 - for WebRTC over TCP
* 3478/UDP - for TURN/UDP
* 50000-60000/UDP - for WebRTC over UDP
* 1935 - for RTMP Ingress
* 7885/UDP - for WHIP Ingress WebRTC

Server URL: wss://video.recontext.online
RTMP Ingress URL: rtmp://video.recontext.online/x
WHIP Ingress URL: https://ingress.recontext.online/w
API Key: APIBj3yrXtyPRNq
API Secret: 2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA

Here's a test token generated with your keys: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3OTg2Njg1NjIsImlzcyI6IkFQSUJqM3lyWHR5UFJOcSIsIm5hbWUiOiJUZXN0IFVzZXIiLCJuYmYiOjE3NjI2Njg1NjIsInN1YiI6InRlc3QtdXNlciIsInZpZGVvIjp7InJvb20iOiJteS1maXJzdC1yb29tIiwicm9vbUpvaW4iOnRydWV9fQ.4gl1QkucFnwPcJzsEWKaQXj7oXBROITDJ4snk8xWfyk

