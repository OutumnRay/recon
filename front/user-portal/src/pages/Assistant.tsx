import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { LuBot, LuSend, LuFileText, LuSearch, LuBookOpen, LuX } from 'react-icons/lu';
import './Assistant.css';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  referencedFiles?: Array<{ id: string; title: string }>;
}

interface FileOption {
  id: string;
  title: string;
  original_name: string;
}

type AssistantMode = 'summary' | 'definitions' | 'find_videos' | 'chat';

type UiState =
  | { kind: 'idle' }
  | { kind: 'file_select'; mode: 'summary' | 'definitions' }
  | { kind: 'find_videos' };

export const Assistant: React.FC = () => {
  const { t } = useTranslation();
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputText, setInputText] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [files, setFiles] = useState<FileOption[]>([]);
  const [selectedFileId, setSelectedFileId] = useState('');
  const [uiState, setUiState] = useState<UiState>({ kind: 'idle' });
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const chatInputRef = useRef<HTMLTextAreaElement>(null);
  const findInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    document.title = `Recontext - ${t('assistant.title')}`;
    fetchFiles();
  }, [t]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isLoading]);

  const fetchFiles = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const resp = await fetch('/api/v1/files?status=completed&page_size=100', {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (resp.ok) {
        const data = await resp.json();
        setFiles(data.items || []);
      }
    } catch {
      // silently fail — user can still use chat mode
    }
  };

  const addUserMessage = (text: string): string => {
    const id = `u-${Date.now()}`;
    setMessages(prev => [...prev, { id, role: 'user', content: text }]);
    return id;
  };

  const sendRequest = useCallback(
    async (message: string, mode: AssistantMode, fileId?: string) => {
      addUserMessage(message);
      setIsLoading(true);
      try {
        const token = localStorage.getItem('token') || sessionStorage.getItem('token');
        const resp = await fetch('/api/v1/assistant/chat', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({ message, mode, file_id: fileId }),
        });

        const data = await resp.json();
        if (!resp.ok) {
          throw new Error(data.error || `HTTP ${resp.status}`);
        }

        setMessages(prev => [
          ...prev,
          {
            id: `a-${Date.now()}`,
            role: 'assistant',
            content: data.response,
            referencedFiles: data.referenced_files,
          },
        ]);
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : String(err);
        setMessages(prev => [
          ...prev,
          {
            id: `a-${Date.now()}`,
            role: 'assistant',
            content: t('assistant.errors.requestFailed') + (msg ? `\n${msg}` : ''),
          },
        ]);
      } finally {
        setIsLoading(false);
      }
    },
    [t]
  );

  const handleSendChat = () => {
    const text = inputText.trim();
    if (!text || isLoading) return;
    setInputText('');
    sendRequest(text, 'chat');
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendChat();
    }
  };

  const handleFileSelectorConfirm = () => {
    if (uiState.kind !== 'file_select' || !selectedFileId) return;
    const mode = uiState.mode;
    const file = files.find(f => f.id === selectedFileId);
    if (!file) return;

    const fileName = file.title || file.original_name;
    const message =
      mode === 'summary'
        ? t('assistant.messages.summaryRequest', { title: fileName })
        : t('assistant.messages.definitionsRequest', { title: fileName });

    setUiState({ kind: 'idle' });
    setSelectedFileId('');
    sendRequest(message, mode, selectedFileId);
  };

  const handleFindVideosSend = () => {
    const text = inputText.trim();
    if (!text || isLoading) return;
    setInputText('');
    setUiState({ kind: 'idle' });
    const message = t('assistant.messages.findVideosRequest', { query: text });
    sendRequest(message, 'find_videos');
  };

  const handleFindVideosKey = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') handleFindVideosSend();
    if (e.key === 'Escape') { setUiState({ kind: 'idle' }); setInputText(''); }
  };

  const cancelPanel = () => {
    setUiState({ kind: 'idle' });
    setSelectedFileId('');
    setInputText('');
  };

  useEffect(() => {
    if (uiState.kind === 'find_videos') {
      setTimeout(() => findInputRef.current?.focus(), 50);
    }
  }, [uiState]);

  const fileLabel = (f: FileOption) => f.title || f.original_name;

  return (
    <div className="assistant-page">
      {/* Quick action cards */}
      <div className="assistant-actions">
        <button
          className="assistant-action-card"
          onClick={() => { setUiState({ kind: 'file_select', mode: 'summary' }); setSelectedFileId(''); }}
          disabled={isLoading}
        >
          <LuFileText className="action-card-icon" />
          <span className="action-card-label">{t('assistant.actions.summary')}</span>
          <span className="action-card-hint">{t('assistant.actions.summaryHint')}</span>
        </button>

        <button
          className="assistant-action-card"
          onClick={() => { setUiState({ kind: 'file_select', mode: 'definitions' }); setSelectedFileId(''); }}
          disabled={isLoading}
        >
          <LuBookOpen className="action-card-icon" />
          <span className="action-card-label">{t('assistant.actions.definitions')}</span>
          <span className="action-card-hint">{t('assistant.actions.definitionsHint')}</span>
        </button>

        <button
          className="assistant-action-card"
          onClick={() => { setUiState({ kind: 'find_videos' }); setInputText(''); }}
          disabled={isLoading}
        >
          <LuSearch className="action-card-icon" />
          <span className="action-card-label">{t('assistant.actions.findVideos')}</span>
          <span className="action-card-hint">{t('assistant.actions.findVideosHint')}</span>
        </button>
      </div>

      {/* Inline panel: file selector */}
      {uiState.kind === 'file_select' && (
        <div className="assistant-panel surface-card">
          <div className="assistant-panel-header">
            <span>{t('assistant.fileSelector.prompt')}</span>
            <button className="panel-close-btn" onClick={cancelPanel}><LuX /></button>
          </div>
          {files.length === 0 ? (
            <p className="panel-empty">{t('assistant.fileSelector.noFiles')}</p>
          ) : (
            <select
              className="assistant-file-select"
              value={selectedFileId}
              onChange={e => setSelectedFileId(e.target.value)}
            >
              <option value="">{t('assistant.fileSelector.placeholder')}</option>
              {files.map(f => (
                <option key={f.id} value={f.id}>{fileLabel(f)}</option>
              ))}
            </select>
          )}
          <div className="assistant-panel-footer">
            <button
              className="assistant-btn-primary"
              onClick={handleFileSelectorConfirm}
              disabled={!selectedFileId}
            >
              {t('assistant.fileSelector.confirm')}
            </button>
            <button className="assistant-btn-ghost" onClick={cancelPanel}>
              {t('common.cancel')}
            </button>
          </div>
        </div>
      )}

      {/* Inline panel: find videos */}
      {uiState.kind === 'find_videos' && (
        <div className="assistant-panel surface-card">
          <div className="assistant-panel-header">
            <span>{t('assistant.findVideos.prompt')}</span>
            <button className="panel-close-btn" onClick={cancelPanel}><LuX /></button>
          </div>
          <div className="find-videos-row">
            <input
              ref={findInputRef}
              type="text"
              className="assistant-text-input"
              placeholder={t('assistant.findVideos.placeholder')}
              value={inputText}
              onChange={e => setInputText(e.target.value)}
              onKeyDown={handleFindVideosKey}
              disabled={isLoading}
            />
            <button
              className="assistant-send-btn"
              onClick={handleFindVideosSend}
              disabled={!inputText.trim() || isLoading}
            >
              <LuSearch />
            </button>
          </div>
        </div>
      )}

      {/* Chat history */}
      <div className="assistant-chat">
        {messages.length === 0 && (
          <div className="assistant-welcome">
            <LuBot className="welcome-icon" />
            <h3 className="welcome-title">{t('assistant.welcome.title')}</h3>
            <p className="welcome-description">{t('assistant.welcome.description')}</p>
          </div>
        )}

        {messages.map(msg => (
          <div key={msg.id} className={`msg-row msg-row--${msg.role}`}>
            {msg.role === 'assistant' && (
              <div className="msg-avatar"><LuBot /></div>
            )}
            <div className={`msg-bubble msg-bubble--${msg.role}`}>
              <p className="msg-text">{msg.content}</p>
              {msg.referencedFiles && msg.referencedFiles.length > 0 && (
                <div className="msg-refs">
                  <span className="refs-label">{t('assistant.referencedFiles')}:</span>
                  <div className="refs-chips">
                    {msg.referencedFiles.map(f => (
                      <span key={f.id} className="ref-chip">{f.title}</span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        ))}

        {isLoading && (
          <div className="msg-row msg-row--assistant">
            <div className="msg-avatar"><LuBot /></div>
            <div className="msg-bubble msg-bubble--assistant msg-bubble--typing">
              <span className="dot" /><span className="dot" /><span className="dot" />
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Chat input — only shown in idle state */}
      {uiState.kind === 'idle' && (
        <div className="assistant-input-bar">
          <textarea
            ref={chatInputRef}
            className="assistant-textarea"
            placeholder={t('assistant.inputPlaceholder')}
            value={inputText}
            onChange={e => setInputText(e.target.value)}
            onKeyDown={handleKeyDown}
            rows={2}
            disabled={isLoading}
          />
          <button
            className="assistant-send-btn"
            onClick={handleSendChat}
            disabled={!inputText.trim() || isLoading}
            aria-label={t('assistant.send')}
          >
            <LuSend />
          </button>
        </div>
      )}
    </div>
  );
};

export default Assistant;
