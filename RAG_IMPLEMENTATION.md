# RAG Implementation Guide

## ✅ COMPLETED

### 1. Database Layer (100% Complete)
**Files Modified:**
- `pkg/database/database.go` - Added RAG Users group and document_chunks table with vector embeddings

**Features:**
- ✅ Created "RAG Users" group with search permissions
- ✅ Created `document_chunks` table with pgvector support
- ✅ Added ivfflat index for fast vector similarity search
- ✅ Enabled pgvector extension

**Permissions for RAG Users:**
```json
{
  "rag": {"actions": ["read", "search"], "scope": "all"},
  "transcriptions": {"actions": ["read"], "scope": "all"}
}
```

### 2. Data Models (100% Complete)
**File Created:** `internal/models/rag.go`

**Models:**
- ✅ `DocumentChunk` - Stores text chunks with embeddings
- ✅ `RAGSearchRequest` - Search query parameters
- ✅ `RAGSearchResult` - Search result format
- ✅ `RAGSearchResponse` - API response format
- ✅ `RAGStatusResponse` - System status

### 3. Database Repository (100% Complete)
**File Created:** `pkg/database/rag_repository.go`

**Functions:**
- ✅ `CreateDocumentChunk()` - Store chunks with embeddings
- ✅ `SearchSimilarChunks()` - Vector similarity search
- ✅ `GetRAGStatus()` - System statistics
- ✅ `DeleteChunksByFileID()` - Clean up chunks
- ✅ `CheckUserHasRAGPermission()` - Permission verification

### 4. Embeddings Service (100% Complete)
**File Created:** `pkg/embeddings/openai.go`

**Features:**
- ✅ OpenAI API integration for text-embedding-3-small
- ✅ `GetEmbedding()` - Generate embeddings from text
- ✅ `ChunkText()` - Split text into manageable chunks
- ✅ Environment variable configuration

## 🚧 TO DO

### 5. Backend API Endpoints (Priority: HIGH)

#### Add RAG Handlers to user-portal/main.go

**Step 1: Update UserPortal struct**
```go
type UserPortal struct {
	config            *config.Config
	logger            *logger.Logger
	jwtManager        *auth.JWTManager
	db                *database.DB
	userRepo          *database.UserRepository
	recordings        map[string]*models.Recording
	prometheusMetrics *metrics.ServiceMetrics
	embeddingsClient  *embeddings.OpenAIEmbeddingsClient  // ADD THIS
}
```

**Step 2: Update NewUserPortal() constructor**
```go
// Initialize embeddings client
embeddingsClient := embeddings.NewOpenAIEmbeddingsClient()

return &UserPortal{
	config:            cfg,
	logger:            log,
	jwtManager:        jwtManager,
	db:                db,
	userRepo:          userRepo,
	recordings:        make(map[string]*models.Recording),
	prometheusMetrics: metrics.NewServiceMetrics("user_portal"),
	embeddingsClient:  embeddingsClient,  // ADD THIS
}, nil
```

**Step 3: Add import statement**
```go
import (
	...
	"Recontext.online/pkg/embeddings"  // ADD THIS
)
```

**Step 4: Add RAG handler functions** (after checkFilePermissionHandler, around line 532):

```go
// RAG Search Handler
func (up *UserPortal) ragSearchHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Check RAG permission
	hasPermission, err := up.db.CheckUserHasRAGPermission(claims.UserID)
	if err != nil || !hasPermission {
		up.respondWithError(w, http.StatusForbidden, "You don't have permission to use RAG search", "")
		return
	}

	var req models.RAGSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Query == "" {
		up.respondWithError(w, http.StatusBadRequest, "Query is required", "")
		return
	}

	// Set defaults
	if req.TopK <= 0 {
		req.TopK = 5
	}
	if req.Threshold <= 0 {
		req.Threshold = 0.7
	}

	// Generate embedding for query
	queryEmbedding, err := up.embeddingsClient.GetEmbedding(req.Query)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate query embedding", err.Error())
		return
	}

	// Search for similar chunks
	results, err := up.db.SearchSimilarChunks(queryEmbedding, req.TopK, req.Threshold)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to search", err.Error())
		return
	}

	response := models.RAGSearchResponse{
		Query:   req.Query,
		Results: results,
		Count:   len(results),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RAG Permission Check Handler
func (up *UserPortal) checkRAGPermissionHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	hasPermission, err := up.db.CheckUserHasRAGPermission(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to check permission", err.Error())
		return
	}

	response := map[string]bool{
		"hasPermission": hasPermission,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RAG Status Handler
func (up *UserPortal) ragStatusHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Check RAG permission
	hasPermission, err := up.db.CheckUserHasRAGPermission(claims.UserID)
	if err != nil || !hasPermission {
		up.respondWithError(w, http.StatusForbidden, "You don't have permission to access RAG status", "")
		return
	}

	status, err := up.db.GetRAGStatus()
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get RAG status", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
```

**Step 5: Add routes in setupRoutes()** (after file routes, around line 633):

```go
// RAG endpoints
mux.Handle("/api/v1/rag/search", chainMiddleware(
	http.HandlerFunc(up.ragSearchHandler),
	authMiddleware,
))

mux.Handle("/api/v1/rag/permission", chainMiddleware(
	http.HandlerFunc(up.checkRAGPermissionHandler),
	authMiddleware,
))

mux.Handle("/api/v1/rag/status", chainMiddleware(
	http.HandlerFunc(up.ragStatusHandler),
	authMiddleware,
))
```

### 6. Managing Portal UI (Priority: HIGH)

**File to Edit:** `front/managing-portal/src/components/UserManagement.tsx`

Add "RAG User" checkbox column next to "File Uploader" checkbox:

```typescript
// In the handler function (after handleToggleFileUploader):
const handleToggleRAGUser = async (user: User) => {
  try {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    const isRAGUser = user.groups?.includes('group-rag-users') || false;
    const endpoint = isRAGUser ? '/api/v1/groups/remove-user' : '/api/v1/groups/add-user';

    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        user_id: user.id,
        group_id: 'group-rag-users'
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to ${isRAGUser ? 'remove user from' : 'add user to'} RAG Users group`);
    }

    fetchUsers();
  } catch (err) {
    alert(err instanceof Error ? err.message : 'Failed to update RAG user permission');
  }
};

// In the table header (add after "File Uploader"):
<th>RAG User</th>

// In the table body (add after File Uploader checkbox):
<td>
  <input
    type="checkbox"
    checked={user.groups?.includes('group-rag-users') || false}
    onChange={() => handleToggleRAGUser(user)}
    className="rag-user-checkbox"
  />
</td>
```

### 7. User Portal RAG Search UI (Priority: HIGH)

**Create File:** `front/user-portal/src/components/RAGSearch.tsx`

```typescript
import React, { useState } from 'react';
import './RAGSearch.css';

interface RAGResult {
  chunk_id: string;
  file_id: string;
  file_name: string;
  chunk_text: string;
  chunk_index: number;
  similarity: number;
  uploaded_at: string;
}

export const RAGSearch: React.FC = () => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<RAGResult[]>([]);
  const [searching, setSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setSearching(true);
    setError(null);

    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/rag/search', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          query: query,
          top_k: 5,
          threshold: 0.7
        })
      });

      if (!response.ok) {
        throw new Error('Search failed');
      }

      const data = await response.json();
      setResults(data.results || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed');
    } finally {
      setSearching(false);
    }
  };

  return (
    <div className="rag-search-container">
      <div className="search-header">
        <h2>Semantic Search</h2>
        <p>Search through transcriptions using natural language</p>
      </div>

      <form onSubmit={handleSearch} className="search-form">
        <div className="search-input-group">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Ask a question or describe what you're looking for..."
            className="search-input"
          />
          <button
            type="submit"
            disabled={searching || !query.trim()}
            className="search-button"
          >
            {searching ? 'Searching...' : 'Search'}
          </button>
        </div>
      </form>

      {error && (
        <div className="error-message">
          {error}
        </div>
      )}

      {results.length > 0 && (
        <div className="results-container">
          <h3>{results.length} results found</h3>
          {results.map((result, index) => (
            <div key={result.chunk_id} className="result-card">
              <div className="result-header">
                <span className="result-filename">{result.file_name}</span>
                <span className="result-similarity">
                  {Math.round(result.similarity * 100)}% match
                </span>
              </div>
              <p className="result-text">{result.chunk_text}</p>
              <div className="result-footer">
                <span>Uploaded: {new Date(result.uploaded_at).toLocaleDateString()}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {results.length === 0 && !searching && query && (
        <div className="no-results">
          No results found. Try a different query.
        </div>
      )}
    </div>
  );
};
```

**Create CSS:** `front/user-portal/src/components/RAGSearch.css`

**Update Dashboard.tsx** - Add conditional RAG Search tab:

```typescript
const [hasRAGPermission, setHasRAGPermission] = useState(false);

useEffect(() => {
  checkRAGPermission();
}, []);

const checkRAGPermission = async () => {
  try {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    const response = await fetch('/api/v1/rag/permission', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    if (response.ok) {
      const data = await response.json();
      setHasRAGPermission(data.hasPermission || false);
    }
  } catch (err) {
    console.error('Failed to check RAG permission:', err);
  }
};

// In navigation:
{hasRAGPermission && (
  <NavLink to="/dashboard/rag-search">
    <LuSearch /> RAG Search
  </NavLink>
)}
```

## 🔧 ENVIRONMENT VARIABLES

Add to your `.env` or docker-compose:

```env
OPENAI_API_KEY=sk-your-api-key-here
OPENAI_EMBEDDING_MODEL=text-embedding-3-small
```

## 📊 DATABASE SCHEMA

### document_chunks table
| Column | Type | Description |
|--------|------|-------------|
| id | VARCHAR(255) | Primary key |
| file_id | VARCHAR(255) | FK to uploaded_files |
| transcription_id | VARCHAR(255) | FK to file_transcriptions |
| chunk_text | TEXT | Text chunk |
| chunk_index | INTEGER | Chunk position |
| embedding | vector(1536) | OpenAI embedding |
| metadata | JSONB | Additional metadata |
| created_at | TIMESTAMP | Creation time |

### Indexes
- ivfflat index on embedding for fast vector similarity search
- B-tree indexes on file_id and transcription_id

## 🎯 TESTING CHECKLIST

- [ ] Add user to "RAG Users" group via managing portal
- [ ] Verify "RAG Search" tab appears in user portal
- [ ] Upload and transcribe a file
- [ ] Generate embeddings for transcription chunks
- [ ] Perform semantic search
- [ ] Verify results are relevant
- [ ] Remove user from RAG group
- [ ] Verify RAG Search tab disappears

## 📝 NEXT STEPS SUMMARY

1. ✅ Database schema created
2. ✅ Models created
3. ✅ Repository functions created
4. ✅ Embeddings service created
5. **TODO:** Add RAG handlers to user-portal/main.go
6. **TODO:** Add RAG User checkbox in managing portal
7. **TODO:** Create RAG Search UI component
8. **TODO:** Create worker to process transcriptions and generate embeddings
9. **TODO:** Set up OpenAI API key

The backend foundation is complete. Implementation requires:
- Adding handlers and routes to user portal
- UI updates in both portals
- Worker process to generate embeddings from transcriptions
