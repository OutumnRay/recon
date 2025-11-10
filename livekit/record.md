curl -X POST video/twirp/livekit.RoomService/CreateRoom \
-H "Authorization: Bearer <token-with-roomCreate>" \
-H 'Content-Type: application/json' \
--data-binary @- << EOF
{
"name": "my-room",
"egress": {
"tracks": {
"filepath": "bucket-path/{room_name}-{publisher_identity}-{time}"
"s3": {
"access_key": "",
"secret": "",
"bucket": "mybucket",
"region": "",
}
}
}
}
EOF




lk token create \
--api-key APIBj3yrXtyPRNq \
--api-secret <SECRET> \
--identity <NAME> \
--room <ROOM_NAME> \
--join \
--valid-for 1h