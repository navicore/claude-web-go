class ChatApp {
    constructor() {
        this.sessionId = this.getOrCreateSessionId();
        this.contextWindowSize = 20;
        this.messages = [];
        this.contextWindow = [];
        
        this.initializeElements();
        this.loadFromLocalStorage();
        this.bindEvents();
        this.renderMessages();
    }
    
    initializeElements() {
        this.messagesEl = document.getElementById('messages');
        this.inputEl = document.getElementById('message-input');
        this.sendBtn = document.getElementById('send-button');
        this.clearBtn = document.getElementById('clear-history');
        this.sessionIdEl = document.getElementById('session-id');
        this.contextSizeEl = document.getElementById('context-size');
        this.contextCountEl = document.getElementById('context-count');
        
        this.sessionIdEl.textContent = `Session: ${this.sessionId.slice(0, 8)}...`;
        this.contextSizeEl.value = this.contextWindowSize;
    }
    
    bindEvents() {
        this.sendBtn.addEventListener('click', () => this.sendMessage());
        this.inputEl.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });
        
        this.clearBtn.addEventListener('click', () => this.clearHistory());
        this.contextSizeEl.addEventListener('change', (e) => {
            this.contextWindowSize = parseInt(e.target.value);
            this.updateContextWindow();
            this.saveToLocalStorage();
        });
    }
    
    getOrCreateSessionId() {
        const stored = localStorage.getItem('claude-chat-session');
        if (stored) {
            const session = JSON.parse(stored);
            if (session.sessionId) {
                return session.sessionId;
            }
        }
        return this.generateUUID();
    }
    
    generateUUID() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c === 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }
    
    loadFromLocalStorage() {
        const stored = localStorage.getItem('claude-chat-session');
        if (stored) {
            const session = JSON.parse(stored);
            this.messages = session.messages || [];
            this.contextWindowSize = session.contextWindowSize || 20;
            this.contextSizeEl.value = this.contextWindowSize;
        }
        this.updateContextWindow();
    }
    
    saveToLocalStorage() {
        const session = {
            sessionId: this.sessionId,
            messages: this.messages.slice(-200),
            contextWindowSize: this.contextWindowSize,
            lastActive: new Date().toISOString()
        };
        localStorage.setItem('claude-chat-session', JSON.stringify(session));
    }
    
    updateContextWindow() {
        this.contextWindow = this.messages.slice(-this.contextWindowSize);
        this.contextCountEl.textContent = this.contextWindow.length;
        
        document.querySelectorAll('.message').forEach(el => {
            el.classList.remove('in-context');
        });
        
        const contextIds = new Set(this.contextWindow.map(m => m.id));
        document.querySelectorAll('.message').forEach(el => {
            if (contextIds.has(el.dataset.messageId)) {
                el.classList.add('in-context');
            }
        });
    }
    
    async sendMessage() {
        const content = this.inputEl.value.trim();
        if (!content) return;
        
        this.sendBtn.disabled = true;
        this.inputEl.disabled = true;
        
        const userMessage = {
            id: this.generateUUID(),
            role: 'user',
            content: content,
            timestamp: new Date().toISOString()
        };
        
        this.messages.push(userMessage);
        this.renderMessage(userMessage);
        this.inputEl.value = '';
        
        this.showTypingIndicator();
        
        try {
            const response = await fetch('/api/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    message: content,
                    sessionId: this.sessionId,
                    contextWindow: this.contextWindow
                })
            });
            
            const data = await response.json();
            
            if (data.error) {
                throw new Error(data.error);
            }
            
            this.messages.push(data.message);
            this.renderMessage(data.message);
            
        } catch (error) {
            const errorMessage = {
                id: this.generateUUID(),
                role: 'assistant',
                content: `Error: ${error.message}`,
                timestamp: new Date().toISOString()
            };
            this.messages.push(errorMessage);
            this.renderMessage(errorMessage);
        } finally {
            this.hideTypingIndicator();
            this.sendBtn.disabled = false;
            this.inputEl.disabled = false;
            this.inputEl.focus();
            this.updateContextWindow();
            this.saveToLocalStorage();
        }
    }
    
    renderMessages() {
        this.messagesEl.innerHTML = '';
        this.messages.forEach(msg => this.renderMessage(msg));
        this.updateContextWindow();
    }
    
    renderMessage(message) {
        const messageEl = document.createElement('div');
        messageEl.className = `message ${message.role}`;
        messageEl.dataset.messageId = message.id;
        
        const headerEl = document.createElement('div');
        headerEl.className = 'message-header';
        headerEl.innerHTML = `
            <strong>${message.role === 'user' ? 'You' : 'Claude'}</strong>
            <span>${new Date(message.timestamp).toLocaleTimeString()}</span>
        `;
        
        const contentEl = document.createElement('div');
        contentEl.className = 'message-content';
        contentEl.innerHTML = marked.parse(message.content);
        
        messageEl.appendChild(headerEl);
        messageEl.appendChild(contentEl);
        
        if (message.files && message.files.length > 0) {
            const filesEl = this.renderFiles(message.files);
            messageEl.appendChild(filesEl);
        }
        
        this.messagesEl.appendChild(messageEl);
        
        Prism.highlightAllUnder(messageEl);
        
        this.messagesEl.scrollTop = this.messagesEl.scrollHeight;
    }
    
    renderFiles(files) {
        const filesContainer = document.createElement('div');
        filesContainer.className = 'files-container';
        
        files.forEach(file => {
            const fileEl = document.createElement('div');
            fileEl.className = 'file-item';
            
            const fileUrl = `/api/files/${this.sessionId}/${file.name}`;
            
            if (file.mimeType.startsWith('image/')) {
                const preview = document.createElement('div');
                preview.className = 'file-preview';
                preview.innerHTML = `<img src="${fileUrl}" alt="${file.name}">`;
                fileEl.appendChild(preview);
            }
            
            const link = document.createElement('a');
            link.href = fileUrl;
            link.download = file.name;
            link.className = 'download-link';
            link.textContent = `Download ${file.name}`;
            fileEl.appendChild(link);
            
            filesContainer.appendChild(fileEl);
        });
        
        return filesContainer;
    }
    
    showTypingIndicator() {
        const indicator = document.createElement('div');
        indicator.id = 'typing-indicator';
        indicator.className = 'message assistant';
        indicator.innerHTML = '<div class="loading"></div> Claude is thinking...';
        this.messagesEl.appendChild(indicator);
        this.messagesEl.scrollTop = this.messagesEl.scrollHeight;
    }
    
    hideTypingIndicator() {
        const indicator = document.getElementById('typing-indicator');
        if (indicator) {
            indicator.remove();
        }
    }
    
    clearHistory() {
        if (confirm('Are you sure you want to clear the conversation history?')) {
            this.messages = [];
            this.contextWindow = [];
            this.sessionId = this.generateUUID();
            this.sessionIdEl.textContent = `Session: ${this.sessionId.slice(0, 8)}...`;
            this.saveToLocalStorage();
            this.renderMessages();
        }
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new ChatApp();
});