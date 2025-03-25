<template>
  <div class="chat-container">
    <!-- Sidebar con le chat -->
    <div class="sidebar">
      <div class="sidebar-header">
        <h2>Chats</h2>
        <button @click="logout" class="logout-button">Logout</button>
      </div>
      <div class="search-bar">
        <div class="search-input-wrapper">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchquery"
            type="text"
            placeholder="Search user..."
            class="search-input"
          />
        </div>
      </div>
      <ul class="chat-list" v-if="chats && chats.length > 0">
        <li
          v-for="chat in chats"
          :key="chat.conversationId"
          @click="selectChat(chat)"
          :class="{ active: chat.conversationId === currentChat?.conversationId }"
        >
          <div class="chat-item">
            <div class="chat-avatar">💬</div>
            <div class="chat-info">
              <strong>{{ chat.nameChat }}</strong>
              <p v-if="chat.lastMessage">{{ truncatedMessage(chat.lastMessage.content) }}</p>
            </div>
          </div>
        </li>
      </ul>
      <p v-else class="no-chats">You have no conversations yet. Start one now!</p>
      <div class="new-chat-wrapper">
        <button @click="startNewChat" class="new-chat-button">✚ New Chat</button>
      </div>
    </div>

    <!-- Area principale della chat -->
    <div class="chat-window" v-if="currentChat">
      <header class="chat-header">
        <h2>{{ currentChat.nameChat }}</h2>
      </header>
      <div class="messages">
        <div
          v-for="(message, index_in_array) in currentChat.messages"
          :key="message.message_id"
          :class="['message', message.username === currentUser ? 'sent' : 'received']"
          :ref="index_in_array === currentChat.messages.length - 1 ? 'lastMessage' : null"
          @mouseleave="hoveredMessage = null"
        >
          <div class="message-header">
            <div class="message-text">
              <p v-if="currentChat.chatType === 'group_chat' && message.username !== currentUser" class="sender-name">
                <strong>{{ getNickname(message.username) }}</strong>
              </p>
              <p class="message-content">{{ message.content }}</p>
            </div>
            <div class="message-options-wrapper">
              <div class="message-options" @click="toggleOptionsMenu(message.message_id)">⋮</div>
              <div v-if="selectedMessageOptions === message.message_id" class="dropdown-menu">
                <p @click="forwardMessage(message)">📤 Inoltra</p>
                <p @click="showEmoji(message)">☺️​ Reazione</p>
                <p v-if="message.username === currentUser" @click="deleteMessage(message)">🗑️ Elimina</p>
              </div>
            </div>
          </div>

          <div v-if="reactionMessageId === message.message_id" class="emoji-op">
            <span
              v-for="emoji in emojiOptions"
              :key="emoji"
              class="emoji-option"
              @click="addReaction(message, emoji)"
            >
              {{ emoji }}
            </span>
          </div>
          <div class="message-reactions" v-if="message.comments && message.comments.length > 0">
            <span v-for="comment in message.comments" :key="comment.username + comment.emojiCode">
              {{ comment.emojiCode }}
            </span>
          </div>
          <span class="message-time">{{ formatTime(message.timestamp) }}</span>
          <span class="status-message" v-if="message.username === currentUser">
            <i v-if="!message.is_read && message.is_delivered" class="fas fa-check is_delivered"></i>
            <i v-else-if="!message.is_read" class="fas fa-check"></i>
            <i v-else class="fas fa-check-double letto"></i>
          </span>
        </div>
      </div>
      <div class="input-area">
        <input type="text" v-model="newMessage" @keyup.enter="sendMessage" placeholder="Type a message..." />
        <button @click="sendMessage">Send</button>
      </div>
    </div>

    <div class="chat-window empty" v-else>
      <p>Select a chat to start messaging</p>
    </div>

    <!-- Modal per la selezione degli utenti -->
    <div v-if="showUserSelection" class="modal">
      <div class="modal-content full-modal">
        <div class="modal-left">
          <h3>Choose Participants</h3>
          <ul class="user-list">
            <li v-for="user in users" :key="user.user_id">
              <label class="user-label">
                <input type="checkbox" :value="user.nickname" v-model="selectedUsers" class="user-checkbox" />
                <img :src="user.profile_image" alt="User avatar" class="user-avatar" />
                <span class="user-name">{{ user.nickname }}</span>
              </label>
            </li>
          </ul>
        </div>

        <div class="modal-right">
          <h3>Create New Chat</h3>
          <div v-if="selectedUsers.length > 1">
            <input
              type="text"
              v-model="groupName"
              placeholder="Enter group name"
              class="start-message-input"
            />
            <input
              type="text"
              v-model="groupImage"
              placeholder="Group image URL (optional)"
              class="start-message-input"
            />
          </div>

          <textarea
            v-model="startMessageText"
            placeholder="Type your start message..."
            rows="4"
            class="start-message-input"
          ></textarea>

          <div class="modal-actions">
            <button @click="createConversation">Create</button>
            <button @click="cancelSelection">Cancel</button>
          </div>

          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      chats: [],
      currentChat: null,
      newMessage: "",
      showUserSelection: false,
      users: [],
      selectedUsers: [],
      groupName: "",
      errorMessage: "",
      searchquery: "",
      messageArray: [],
      selectedMessageOptions: null,
      hoveredMessage: null,
      reactionMessageId: null,
      emojiOptions: ["👍", "😂​", "❤️​"],
    };
  },
  created() {
    this.token = Number(localStorage.getItem("token"));
    this.currentUser = localStorage.getItem("username") || "You";
    if (!this.token || this.token <= 0) {
      this.$router.push("/login");
    } else {
      this.fetchChats();
    }
  },
  methods: {
    async fetchMessageHistory(conversation_id) {
      try {
        const token = localStorage.getItem("token");
        if (!token) throw new Error("Non autorizzato");

        const response = await fetch(`${__API_URL__}/conversation/${conversation_id}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          throw new Error("Errore HTTP: " + response.status);
        }

        const data = await response.json();
        const chatFromSidebar = this.chats.find(c => c.conversationId === conversation_id);
        console.log(data.messages);
        this.currentChat = {
          ...chatFromSidebar,
          chatType: chatFromSidebar?.ChatType,
          messages: data.messages.map(msg => {
            const isMyMessage = msg.username === this.currentUser;
            let is_read = false;
            let is_delivered = false;
            if (isMyMessage && msg.read_status && msg.read_status.length > 0) {
              const otherUsers = msg.read_status.filter(r => r.user_id !== this.currentUserId);

              is_read = otherUsers.length > 0 && otherUsers.every(r => r.is_read === 1 || r.is_read === true);
              is_delivered = otherUsers.length > 0 && otherUsers.every(r => r.is_delivered === 1 || r.is_delivered === true);
            }

            return {
              ...msg,
              is_read,
              is_delivered
            };
          }),
          users: data.utenti?.users || []
        };

        this.$nextTick(() => {
          this.scrollToLastMessage();
        });

      } catch (error) {
        console.error("Errore durante il fetch della conversazione:", error);
      }
    },
    async fetchChats() {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non autorizzato");
        }
        const response = await fetch(`${__API_URL__}/conversations`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        if(!response.ok){
          throw new Error("Errore HTTP: " + response.status);
        }
        const data = await response.json();
        this.chats = data.conversation;
      } catch (error) {
        console.error("Errore nel fetch delle chat:", error);
      }
    },
    async fetchUsers() {
      try {
        const token = localStorage.getItem("token");
        const response = await this.$axios.get("/users", {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!response.data || !response.data.users) {
          throw new Error("La chiave 'users' non esiste nella risposta");
        }
        this.users = response.data.users;
        this.users = response.data.users.filter(u => u.nickname !== this.currentUser);
      } catch (error) {
        console.error("Errore nel fetch degli utenti:", error);
      }
    },
    selectChat(chat) {
      this.currentChat = chat
      if (this.currentChat) {
        this.fetchMessageHistory(this.currentChat.conversationId);
      } else {
        console.warn("Nessuna chat trovata con id:", chat.conversationId);
      }

    },
    async sendMessage() {
      if (!this.newMessage.trim()) return;

      try {
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non autorizzato");
        }

        const messagePayload = {
          content: this.newMessage,   // Il testo del messaggio
          media: "text",              // Di default è testo, ma potrebbe essere gif o altro
          image: ""                   // Vuoto per i messaggi di solo testo
        };

        // Chiamata API per inviare il messaggio
        const response = await fetch(`${__API_URL__}/conversation/${this.currentChat.conversationId}`, {
          method: "POST",
          headers: {
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify(messagePayload)
        });

        if (!response.ok) {
          throw new Error(`Errore HTTP: ${response.status}`);
        }
        const data = await response.json();
        this.currentChat.messages.push({
          message_id: data.messageId || Date.now(),
          username: this.currentUser,
          content: this.newMessage,
          timestamp: new Date().toISOString(),
        });
        const chat = this.chats.find(c => c.conversationId === this.currentChat.conversationId);
        if (chat) {
          chat.lastMessage = {
            content: this.newMessage,
            timestamp: new Date().toISOString()
          };

          const index = this.chats.findIndex(c => c.conversationId === this.currentChat.conversationId);
          if(index > -1){
            const[updateChat] = this.chats.splice(index, 1);
            this.chats.unshift(updateChat)
          }
        }
        this.newMessage = "";
        this.$nextTick(() => {
          this.scrollToLastMessageWithSmooth();
        });
      } catch (error) {
        console.error("Errore nell'invio del messaggio:", error);
      }
    },
    startNewChat() {
      this.showUserSelection = true;
      this.fetchUsers();
    },
    async createConversation() {
      if (this.selectedUsers.length === 0) {
        this.errorMessage = "Seleziona almeno un utente";
        return;
      }
      if (!this.startMessageText.trim()) {
        this.errorMessage = "Devi scrivere un messaggio iniziale";
        return;
      }


      const nSelected = this.selectedUsers.length;
      let chatTypeValue = null;
      let groupNameValue = '';

      // Se viene selezionato 1 utente: private_chat, altrimenti group_chat
      if (nSelected === 1) {
        chatTypeValue = { ChatType: 'private_chat' };
      } else if (nSelected > 1) {
        if (!this.groupName.trim()) {
          this.errorMessage = "Per un gruppo è necessario un nome";
          return;
        }
        chatTypeValue = { ChatType: 'group_chat' };
        groupNameValue = this.groupName;
      }

      const conversationRequest = {
        chatType: chatTypeValue,                        // Tipo di chat (private_chat, group_chat)
        groupName: groupNameValue,                      // Valido solo per group chat
        imageGroup: nSelected > 1 ? "https://cdn.raceroster.com/assets/images/team-placeholder.png" : "",
        usersname: this.selectedUsers,                // Array degli utenti selezionati
        startMessage: {
          media: "text",             	 // Tipo di media (text, gif, gif_with_text)
          content: this.startMessageText.trim(),
          image: "",
        }
      };

      try {
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }
        const response = await fetch(`${__API_URL__}/conversations`, {
          method: "POST",
          headers: {
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify(conversationRequest)
        });

        const data = await response.json();
        if (!data || !data.ConversationId) {
          throw new Error("La risposta API non contiene un ConversationId valido");
        }

        this.showUserSelection = false;
        this.selectedUsers = [];
        this.groupName = "";
        await this.fetchChats();
        const newChat = this.chats.find(c => c.conversationId === data.ConversationId);
        if(newChat){
          this.selectChat(newChat);
          return;
        }

      } catch (error) {
        this.errorMessage = "Errore durante la creazione della chat: " + error.message;
      }
    },
    async deleteMessage(message){
      try{
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }

        const response = await fetch(`${__API_URL__}/conversation/${this.currentChat.conversationId}/messages/${message.message_id}`, {
          method: "DELETE",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          throw new Error("Errore HTTP: " + response.status);
        }

        this.currentChat.messages = this.currentChat.messages.filter(m => m.message_id !== message.message_id);
        if(this.currentChat.messages.length === 0){
          this.chats = this.chats.filter(c => c.conversationId !== this.currentChat.conversationId);
          this.currentChat = null;
          return;
        }
        const chat = this.chats.find(c => c.conversationId === this.currentChat.conversationId);
        if (chat && chat.lastMessage?.content === message.content) {
          const lastM = this.currentChat.messages[this.currentChat.messages.length - 1];
          chat.lastMessage = lastM ? { content: lastM.content, timestamp: lastM.timestamp } : null;
        }

        this.selectedMessageOptions = null;
      } catch (error) {
        console.error("Errore durante l'eliminazione del messaggio:", error);
      }
    },
    async addReaction(message, emoji){
      try{
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }

        const exist = message.comments?.find(c => c.username === this.currentUser);
        if(exist && exist.emojiCode === emoji){
          return this.removeReaction(message);
        }

        const response = await fetch(`${__API_URL__}/conversation/${this.currentChat.conversationId}/messages/${message.message_id}/comment`, {
          method: "PUT",
          headers: {
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({ emojiCode: emoji })
        });

        if (!response.ok) {
          throw new Error("Errore HTTP: " + response.status);
        }

        if(!message.comments){
          message.comments = [];
        }

        if(exist){
          exist.emojiCode = emoji;
        } else {
          message.comments.push({ username: this.currentUser, emojiCode: emoji});
        }

        this.reactionMessageId = null;
        this.selectedMessageOptions = null;
      } catch (err) {
        console.error("Errore durante l'aggiunta dell'emoji: ", err);
      }
    },
    async removeReaction(message){
      try{
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }

        const response = await fetch(`${__API_URL__}/conversation/${this.currentChat.conversationId}/messages/${message.message_id}/comment`, {
          method: "DELETE",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          throw new Error("Errore HTTP: " + response.status);
        }

        if(message.comments){
          message.comments = message.comments.filter(c => c.username !== this.currentUser);
        }
      

      this.reactionMessageId = null;
      this.selectedMessageOptions = null;
      } catch (err) {
        console.error("Errore durante la rimozione dell'emoji: ", err);
      }
    },
    cancelSelection() {
      this.showUserSelection = false;
      this.selectedUsers = [];
      this.groupName = "";
      this.groupImage = "";
      this.startMessageText = "";
      this.errorMessage = "";
    },
    logout() {
      localStorage.removeItem("token");
      this.$router.push("/login");
    },
    getNickname(name) {
      const lower = String(name).toLowerCase();
      const users = this.currentChat && this.currentChat.users ? this.currentChat.users : [];

      for (let i = 0; i < users.length; i++) {
        const u = users[i];
        const nick = u.nickname || "";
        const usern = u.username || "";
        const uid = String(u.user_id || "");

        if (nick.toLowerCase() === lower || usern.toLowerCase() === lower || uid === lower) {
          return nick || usern || name;
        }
      }

      return name;
    },

    truncatedMessage(msg) {
      if(!msg) return "";
      const maxChars = 30;
      return msg.length > maxChars ? msg.slice(0, maxChars) + "..." : msg;
    },
    formatTime(timestamp) {
      const date = new Date(timestamp);
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    },
    scrollToLastMessage() {
      const lastMessage = this.$refs.lastMessage;
      if (lastMessage && lastMessage.length > 0) {
        lastMessage[lastMessage.length - 1].scrollIntoView({ behavior: 'auto' });
      }
    },
    scrollToLastMessageWithSmooth() {
      const lastMessage = this.$refs.lastMessage;
      if (lastMessage && lastMessage.length > 0) {
        lastMessage[lastMessage.length - 1].scrollIntoView({ behavior: 'smooth' });
      }
    },
    toggleOptionsMenu(id){
      this.selectedMessageOptions = this.selectedMessageOptions === id ? null : id;
    },
    forwardMessage(message) {
    // Apri modale o selezione destinazione
    console.log("Inoltro", message);
    this.selectedMessageOptions = null;
    },
    showEmoji(message) {
      this.reactionMessageId = message.message_id;
    },
  },

  mounted() {
    this.fetchChats();
  },
};
</script>

<style scoped>
/* Contenitore principale */
.chat-container {
  display: flex;
  height: 100vh;
  background: #f5f5f5;
}

/* Sidebar */
.sidebar {
  width: 25%;
  background: #2c3e50;
  color: white;
  display: flex;
  flex-direction: column;
  padding: 10px;
  flex-shrink: 0;
}

.sidebar ul{
  overflow-y: auto;
}

.new-chat-wrapper {
  position: absolute;
  bottom: 10px;
  right: 10px;
  display: flex;
  justify-content: flex-end;
}

.new-chat-button {
  background: #1abc9c;
  border: none;
  padding: 10px;
  color: rgb(255, 255, 255);
  cursor: pointer;
  border-radius: 5px;
  transition: background 0.3s ease, box-shadow 0.3s ease;
}

.new-chat-button:hover {
  background: #16a085;
  box-shadow: 0 0 15px rgba(26, 188, 156, 0.6);
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.sidebar-header h2 {
  margin: 0;
  font-size: 1.5em;
}

.sidebar-header button {
  background: #1abc9c;
  border: none;
  padding: 5px 10px;
  color: white;
  cursor: pointer;
  border-radius: 5px;
}

/* Lista utenti */
.user-list {
  list-style: none;
  padding: 0;
  margin: 10px 0;
  max-height: 300px;
  overflow-y: auto;
}

.user-item {
  padding: 8px;
  border-radius: 8px;
  transition: background 0.2s ease-in-out;
}

.user-item:hover {
  background: #f1f1f1;
}

.user-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  width: 100%;
}

.user-checkbox {
  width: 18px;
  height: 18px;
  accent-color: #007bff;
}


/* Foto utente */
.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.user-name {
  font-weight: 500;
  font-size: 15px;
  color: #333;
}


.chat-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.chat-list ul{
  overflow-y: auto;
}

.chat-list li {
  padding: 10px;
  border-bottom: 1px solid #34495e;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: background 0.2s;
}

.chat-list li:hover,
.chat-list li.active {
  background: #34495e;
}

.chat-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.chat-avatar {
  background: #1abc9c;
  color: white;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chat-info {
  flex-grow: 1;
}

.chat-info strong {
  display: block;
}

/* Finestra della chat */
.chat-window {
  margin-left: 25%;
  flex-grow: 1;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: white;
  border-left: 1px solid #ddd;
  position: relative;
  z-index: 10;
}

.chat-header {
  position: sticky;
  top: 0;
  background-color: #3498db;
  color: white;
  font-size: 1.2em;
  padding: 10px;
  text-align: center;
  height: 50px;
  display: flex;
  align-items: center;
  z-index: 1;
}

.emoji-op{
  display: flex;
  gap: 10px;
  margin-top: 4px;
  padding: 4px 0;
}

.emoji-option{
  cursor: pointer;
  font-size: 1.2em;
  transition: transform 0.2s ease;
}

.emoji-option:hover{
  transform: scale(1.3);
}

.message-reactions{
  margin-top: 6px;
  font-size: 1.1rem;
}

.messages {
  flex-grow: 1;
  overflow-y: auto;
  padding: 20px;
  height: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 60px;
}

.sender-name {
  font-size: 0.85rem;
  color: #333;
  margin-bottom: 4px;
}

.message-text p {
  margin: 0;
  word-break: break-word;
}

.message-time {
  font-size: 0.75rem;
  color: #ffffff;
  margin-top: 4px;
  display: block;
}

.start-message-input {
  width: 100%;
  padding: 8px;
  resize: vertical;
  margin-top: 10px;
}


.message {
  max-width: 70%;
  padding: 10px;
  border-radius: 10px;
  font-size: 14px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  position: relative;
}

.message-options-wrapper {
  position: relative;
  z-index: 2000;
}

.message.sent .dropdown-menu {
  right: 100%;
  left: auto;
  top: 60px;
  margin-right: 10%;
}

.message.received .dropdown-menu {
  left: 100%;
  right: auto;
  top: 0;
  margin-left: 10%;
}

.message-options {
  cursor: pointer;
  font-weight: bold;
  padding: 0 5px;
}

.dropdown-menu {
  position: absolute;
  right: 0;
  top: 25px;
  background: white;
  border: 1px solid #ccc;
  box-shadow: 0 2px 6px rgba(0,0,0,0.2);
  z-index: 999;
  border-radius: 20%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.dropdown-menu p {
  margin: 0;
  padding: 2px 10px;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.2s;
}


.dropdown-menu p:hover {
  background: #eee;
}


.message.sent {
  align-self: flex-end;
  background: #007bff;
  color: white;
}

.message.received {
  align-self: flex-start;
  background: #f1f0f0;
  color: black;
}

.message-time {
  font-size: 0.75rem;
  margin-top: 5px;
  opacity: 0.7;
}

.sent .message-time {
  color: #fff;
}

.received .message-time {
  color: #333;
}

.status-message {
	font-size: 12px;
	color: rgb(156, 16, 16);
	margin-left: auto;
}

.status-message .letto {
	color: #44dc92;
}

.is_delivered {
  color: yellow;
}
/* Input messaggi */
.input-area {
  margin-left: 25%;
  display: flex;
  padding: 10px;
  gap: 10px;
  background: #f8f9fa;
  border-top: 1px solid #ddd;
  width: 100%;
  position: fixed;
  bottom: 0;
  right: 0;
  box-sizing: border-box;
}

.input-area input {
  margin-left: 25%;
  flex-grow: 1;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 5px;
}

.input-area button {
  background: #007bff;
  border: none;
  color: white;
  padding: 10px 15px;
  margin-left: 10px;
  cursor: pointer;
  border-radius: 5px;
}

/* Quando nessuna chat è selezionata */
.chat-window.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2em;
  color: #777;
}

/* Messaggio se non ci sono chat */
.no-chats {
  color: rgba(255, 255, 255, 0.7);
  text-align: center;
  padding: 20px;
  font-size: 1.1em;
}

/* per la selezione degli utenti */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 999;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 20px;
  width: 90%;
  max-width: 600px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.2);
}

.modal-content h3 {
  margin: 0 0 10px;
  font-size: 22px;
  font-weight: bold;
  text-align: center;
}

.user-list {
  list-style: none;
  padding: 0;
  margin: 0;
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 10px;
  background: #f9f9f9;
}

.user-item {
  display: flex;
  align-items: center;
  padding: 8px;
  border-radius: 8px;
  transition: background 0.2s ease-in-out;
}

.user-item:hover {
  background: #f1f1f1;
}

.user-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  width: 100%;
}

.user-checkbox {
  width: 18px;
  height: 18px;
  accent-color: #007bff;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.user-name {
  font-weight: 500;
  font-size: 15px;
  color: #333;
  flex-grow: 1;
}

.modal-actions {
  display: flex;
  justify-content: space-between;
  gap: 10px;
}

.modal-actions button {
  flex: 1;
  padding: 10px;
  border: none;
  border-radius: 6px;
  font-weight: bold;
  cursor: pointer;
  transition: background 0.3s;
}

.modal-actions button:first-of-type {
  background: #007bff;
  color: white;
}

.modal-actions button:last-of-type {
  background: #ccc;
  color: black;
}

.modal-actions button:hover {
  opacity: 0.9;
}

.start-message-input {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 8px;
  font-size: 15px;
  font-family: inherit;
  transition: border-color 0.2s ease;
  margin-top: 10px;
}

.start-message-input:focus {
  outline: none;
  border-color: #1abc9c;
  box-shadow: 0 0 0 2px #1abc9c33;
}

.error-message {
  color: red;
  font-size: 14px;
  margin-top: 10px;
  text-align: center;
}

.search-bar {
  margin: 0 0 15px 0;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  background-color: #ffffff;
  border-radius: 8px;
  padding: 6px 12px;
  transition: box-shadow 0.3s ease;
}

.search-input-wrapper:focus-within {
  box-shadow: 0 0 0 2px #1abc9c;
}

.search-icon {
  margin-right: 8px;
  color: #2c3e50;
  font-size: 18px;
}

.search-input {
  flex: 1;
  padding: 6px;
  background-color: transparent;
  border: none;
  outline: none;
  color: #2c3e50;
  font-size: 15px;
  font-weight: 400;
}

.logout-button {
  background-color: #e74c3c;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: bold;
  cursor: pointer;
  transition: background-color 0.3s ease, box-shadow 0.3s ease;
}

.logout-button:hover {
  background-color: #c0392b;
  box-shadow: 0 0 10px #e74c3c80;
}
</style>
