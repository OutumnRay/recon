import asyncio
from livekit import api
import httpx
import aiohttp
import ssl
import certifi


async def startComposite():
    # It's recommended to retrieve the API key and secret from environment variables
    # or another secure configuration method rather than hardcoding them.
    # LIVEKIT_API_KEY = "your_api_key"
    # LIVEKIT_API_SECRET = "your_api_secret"

    # Создаем SSL контекст с проверкой и поддержкой SNI (по умолчанию aiohttp уже поддерживает SNI)
    ssl_context = ssl.create_default_context(cafile=certifi.where())
    connector = aiohttp.TCPConnector(ssl=ssl_context)

    # Создаем aiohttp-сессию с этим SSL контекстом
    session = aiohttp.ClientSession(connector=connector)

    req = api.RoomCompositeEgressRequest(
        room_name="my-room",
        layout="speaker",
        custom_base_url="https://video.recontext.online",
        # custom_base_url is for custom layouts, not the LiveKit server URL
        # The LiveKitAPI constructor takes the server URL.
        preset=api.EncodingOptionsPreset.H264_720P_30,
        audio_only=True,
        segment_outputs=[api.SegmentedFileOutput(
            filename_prefix="my-output",
            playlist_name="my-playlist.m3u8",
            live_playlist_name="my-live-playlist.m3u8",
            segment_duration=2,
            s3=api.S3Upload(
                endpoint="https://api.storage.recontext.online",
                bucket="jitsi-recordings",
                region="",
                access_key="minioadmin",
                secret="minioadmin",
                force_path_style=True,
            ),
        )],
    )



    # The custom_base_url should be removed from the request and the LiveKit server URL
    # should be passed to the LiveKitAPI constructor.
    # The URL should generally not include the port unless it's non-standard.
    lkapi = api.LiveKitAPI(
        url="https://video.recontext.online",
        api_key="APIBj3yrXtyPRNq",
        api_secret="2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA",
        session=session
    )

    print("Starting room composite egress...")
    try:
        res = await lkapi.egress.start_room_composite_egress(req)
        print("Egress started successfully:", res)

    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        # It's important to close the client session when you're done.
        await lkapi.aclose()
        await session.close()

async def stop(egress_id):
    # It's recommended to retrieve the API key and secret from environment variables
    # or another secure configuration method rather than hardcoding them.
    # LIVEKIT_API_KEY = "your_api_key"
    # LIVEKIT_API_SECRET = "your_api_secret"

    # Создаем SSL контекст с проверкой и поддержкой SNI (по умолчанию aiohttp уже поддерживает SNI)
    ssl_context = ssl.create_default_context(cafile=certifi.where())
    connector = aiohttp.TCPConnector(ssl=ssl_context)

    # Создаем aiohttp-сессию с этим SSL контекстом
    session = aiohttp.ClientSession(connector=connector)
    req = api.StopEgressRequest(egress_id=egress_id)



    # The custom_base_url should be removed from the request and the LiveKit server URL
    # should be passed to the LiveKitAPI constructor.
    # The URL should generally not include the port unless it's non-standard.
    lkapi = api.LiveKitAPI(
        url="https://video.recontext.online",
        api_key="APIBj3yrXtyPRNq",
        api_secret="2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA",
        session=session
    )

    print("Stoped room composite egress...")
    try:
        res = await lkapi.egress.stop_egress(req)
        print("Egress stoped successfully:", res)

    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        # It's important to close the client session when you're done.
        await lkapi.aclose()
        await session.close()

async def startTrackFile(track_id):
    # It's recommended to retrieve the API key and secret from environment variables
    # or another secure configuration method rather than hardcoding them.
    # LIVEKIT_API_KEY = "your_api_key"
    # LIVEKIT_API_SECRET = "your_api_secret"

    # Создаем SSL контекст с проверкой и поддержкой SNI (по умолчанию aiohttp уже поддерживает SNI)
    ssl_context = ssl.create_default_context(cafile=certifi.where())
    connector = aiohttp.TCPConnector(ssl=ssl_context)

    # Создаем aiohttp-сессию с этим SSL контекстом
    session = aiohttp.ClientSession(connector=connector)

    req = api.TrackEgressRequest(
        room_name="my-room",
        track_id=track_id,
        file=api.DirectFileOutput(
            s3=api.S3Upload(
                endpoint="https://api.storage.recontext.online",
                bucket="jitsi-recordings",
                region="",
                access_key="minioadmin",
                secret="minioadmin",
                force_path_style=True,
            ),
            filepath="{room_name}/{track_id}",
        )

    )

    # The custom_base_url should be removed from the request and the LiveKit server URL
    # should be passed to the LiveKitAPI constructor.
    # The URL should generally not include the port unless it's non-standard.
    lkapi = api.LiveKitAPI(
        url="https://video.recontext.online",
        api_key="APIBj3yrXtyPRNq",
        api_secret="2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA",
        session=session
    )

    print("Starting room composite egress...")
    try:
        res = await lkapi.egress.start_track_egress(req)
        print("Egress started successfully:", res)

    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        # It's important to close the client session when you're done.
        await lkapi.aclose()
        await session.close()

async def startTrackSegmented(track_id, room):
    # It's recommended to retrieve the API key and secret from environment variables
    # or another secure configuration method rather than hardcoding them.
    # LIVEKIT_API_KEY = "your_api_key"
    # LIVEKIT_API_SECRET = "your_api_secret"

    # Создаем SSL контекст с проверкой и поддержкой SNI (по умолчанию aiohttp уже поддерживает SNI)
    ssl_context = ssl.create_default_context(cafile=certifi.where())
    connector = aiohttp.TCPConnector(ssl=ssl_context)

    # Создаем aiohttp-сессию с этим SSL контекстом
    session = aiohttp.ClientSession(connector=connector)



    req = api.TrackCompositeEgressRequest(
        room_name=room,
        audio_track_id=track_id,
        video_track_id=None,
        preset=api.EncodingOptionsPreset.H264_720P_30,
        # a placeholder RTMP output is needed to ensure stream urls can be added to it later
        stream_outputs=None,
        segment_outputs=[api.SegmentedFileOutput(
            filename_prefix=f"{room}/{track_id}",
            playlist_name=f"{room}/{track_id}.m3u8",
            live_playlist_name=f"{room}/live-{track_id}.m3u8",
            segment_duration=20,
            s3=api.S3Upload(
                endpoint="https://api.storage.recontext.online",
                bucket=f"jitsi-recordings",
                region="",
                access_key="minioadmin",
                secret="minioadmin",
                force_path_style=True,
            ),
        )],
    )

    # The custom_base_url should be removed from the request and the LiveKit server URL
    # should be passed to the LiveKitAPI constructor.
    # The URL should generally not include the port unless it's non-standard.
    lkapi = api.LiveKitAPI(
        url="https://video.recontext.online",
        api_key="APIBj3yrXtyPRNq",
        api_secret="2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA",
        session=session
    )

    print("Starting room composite egress...")
    try:
        res = await lkapi.egress.start_track_composite_egress(req)
        print("Egress started successfully:", res)

    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        # It's important to close the client session when you're done.
        await lkapi.aclose()
        await session.close()

if __name__ == "__main__":
    # asyncio.run() starts the event loop and runs the async function until it's complete.
    asyncio.run(stop("EG_RBF6ZkpkbwMN"))
    #asyncio.run(startTrackSegmented("TR_AMc5YKjdKk42aS","my-room"))
    #asyncio.run(startComposite())