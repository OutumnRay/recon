from livekit import api
import os

token = api.AccessToken('APIBj3yrXtyPRNq', '2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA') \
               .with_identity("identity") \
               .with_name("name") \
               .with_grants(api.VideoGrants(
                   room_join=True,
                   room="my-room",
               )).to_jwt()

print(token)