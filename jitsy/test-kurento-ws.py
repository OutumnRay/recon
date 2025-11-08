#!/usr/bin/env python3
"""
Test Kurento WebSocket connection
"""

import asyncio
import json
import websockets

KURENTO_URI = 'ws://kurento:8888/'

async def test_connection():
    """Test WebSocket connection to Kurento"""
    print(f"Testing connection to: {KURENTO_URI}")

    try:
        # Try with subprotocol
        print("\n1. Testing with 'kurento' subprotocol...")
        async with websockets.connect(KURENTO_URI, subprotocols=['kurento']) as ws:
            print(f"✅ Connected with subprotocol!")
            print(f"   Selected subprotocol: {ws.subprotocol}")

            # Send ping request
            request = {
                "id": 1,
                "method": "ping",
                "params": {},
                "jsonrpc": "2.0"
            }
            await ws.send(json.dumps(request))
            response = await ws.recv()
            print(f"   Response: {response}")

    except Exception as e:
        print(f"❌ Failed with subprotocol: {e}")

    try:
        # Try without subprotocol
        print("\n2. Testing WITHOUT subprotocol...")
        async with websockets.connect(KURENTO_URI) as ws:
            print(f"✅ Connected without subprotocol!")

            # Send ping request
            request = {
                "id": 1,
                "method": "ping",
                "params": {},
                "jsonrpc": "2.0"
            }
            await ws.send(json.dumps(request))
            response = await ws.recv()
            print(f"   Response: {response}")

    except Exception as e:
        print(f"❌ Failed without subprotocol: {e}")

    try:
        # Try creating pipeline
        print("\n3. Testing pipeline creation...")
        async with websockets.connect(KURENTO_URI, subprotocols=['kurento']) as ws:
            request = {
                "id": 2,
                "method": "create",
                "params": {
                    "type": "MediaPipeline",
                    "constructorParams": {},
                    "properties": {}
                },
                "jsonrpc": "2.0"
            }

            await ws.send(json.dumps(request))
            response = await ws.recv()
            data = json.loads(response)

            if "result" in data:
                print(f"✅ Pipeline created!")
                print(f"   Pipeline ID: {data['result'].get('value')}")
                print(f"   Session ID: {data['result'].get('sessionId')}")
            else:
                print(f"❌ Error: {data.get('error')}")

    except Exception as e:
        print(f"❌ Failed to create pipeline: {e}")

if __name__ == '__main__':
    asyncio.run(test_connection())
