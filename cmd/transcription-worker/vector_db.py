"""Qdrant vector-DB integration for transcript chunk indexing."""
import logging
import uuid

import config as cfg

logger = logging.getLogger(__name__)

_qdrant = None
_embedder = None


def _get_qdrant():
    global _qdrant
    if _qdrant is None:
        import os
        # Bypass system proxy for localhost connections
        for key in ("no_proxy", "NO_PROXY"):
            existing = os.environ.get(key, "")
            hosts = {h.strip() for h in existing.split(",") if h.strip()}
            hosts.update({"localhost", "127.0.0.1", cfg.QDRANT_HOST})
            os.environ[key] = ",".join(hosts)

        from qdrant_client import QdrantClient
        from qdrant_client.models import Distance, VectorParams

        _qdrant = QdrantClient(
            host=cfg.QDRANT_HOST,
            port=cfg.QDRANT_PORT,
            prefer_grpc=False,   # use HTTP REST, not gRPC
            timeout=30,
        )

        existing = [c.name for c in _qdrant.get_collections().collections]
        if cfg.QDRANT_COLLECTION not in existing:
            _qdrant.create_collection(
                collection_name=cfg.QDRANT_COLLECTION,
                vectors_config=VectorParams(size=384, distance=Distance.COSINE),
            )
            logger.info("Created Qdrant collection: %s", cfg.QDRANT_COLLECTION)
        else:
            logger.debug("Qdrant collection already exists: %s", cfg.QDRANT_COLLECTION)
    return _qdrant


def _get_embedder():
    global _embedder
    if _embedder is None:
        from sentence_transformers import SentenceTransformer
        logger.info("Loading embedding model: %s", cfg.EMBEDDING_MODEL)
        _embedder = SentenceTransformer(cfg.EMBEDDING_MODEL)
        logger.info("Embedding model loaded")
    return _embedder


def _embed(texts: list) -> list:
    vectors = _get_embedder().encode(texts, batch_size=32, show_progress_bar=False)
    return vectors.tolist()


def upsert_to_vector_db(chunks: list) -> bool:
    """
    Upsert transcript chunks into Qdrant.
    Each chunk must contain: text, start, end, speaker, session_id, user_id.

    If QDRANT_ENABLED is false the function is a no-op (returns True).
    """
    if not cfg.QDRANT_ENABLED:
        logger.debug("Qdrant disabled – skipping vector upsert (%d chunks)", len(chunks))
        return True

    if not chunks:
        logger.warning("No chunks to upsert")
        return True

    try:
        from qdrant_client.models import PointStruct

        texts = [c["text"] for c in chunks]
        vectors = _embed(texts)

        points = [
            PointStruct(
                id=str(uuid.uuid4()),
                vector=vector,
                payload={
                    "session_id": chunk.get("session_id", ""),
                    "user_id": chunk.get("user_id", ""),
                    "speaker": chunk.get("speaker", "unknown"),
                    "start": chunk.get("start", 0.0),
                    "end": chunk.get("end", 0.0),
                    "text": chunk["text"],
                },
            )
            for chunk, vector in zip(chunks, vectors)
        ]

        _get_qdrant().upsert(collection_name=cfg.QDRANT_COLLECTION, points=points)
        logger.info("Upserted %d chunks to Qdrant collection '%s'", len(points), cfg.QDRANT_COLLECTION)
        return True

    except Exception as exc:
        logger.error("Qdrant upsert failed: %s", exc)
        return False
