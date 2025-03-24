<template>
  <div class="chat-container">
    <!-- Sidebar con le chat -->
    <div class="sidebar">
      <!-- Parte alta della sidebar -->

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

      <!-- New Chat in basso -->
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
        <div v-for="(message, index_in_array) in currentChat.messages" :key="message.message_id" :class="['message', message.username === currentUser ? 'sent' : 'received']" :ref="index_in_array === currentChat.messages.length - 1 ? 'lastMessage' : null">
          <div class="message-text">
            <!-- Mostra il nome solo se è un gruppo e il messaggio non è mio -->
            <p v-if="currentChat.chatType === 'group_chat' && message.username !== currentUser" class="sender-name">
              <strong>{{ getNickname(message.username) }}</strong>
            </p>
            <!-- Contenuto del messaggio -->
            <p class="message-content">{{ message.content }}</p>
          </div>
          <!-- Timestamp del messaggio -->
          <span class="message-time">{{ formatTime(message.timestamp) }}</span>
          <!-- Status messaggio -->
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

    <!-- Messaggio quando nessuna chat è selezionata -->
    <div class="chat-window empty" v-else>
      <p>Select a chat to start messaging</p>
    </div>

    <!-- Modal per la selezione degli utenti -->
    <div v-if="showUserSelection" class="modal">
      <div class="modal-content">
        <h3>Select Users</h3>
        <ul class="user-list">
          <li v-for="user in users" :key="user.user_id">
            <label class="user-label">
              <input type="checkbox" :value="user.nickname" v-model="selectedUsers" class="user-checkbox" />
              <img :src="user.profile_image" alt="User avatar" class="user-avatar" />
              <span class="user-name">{{ user.nickname }}</span>
            </label>
          </li>
        </ul>
        <div v-if="selectedUsers.length > 1">
          <input type="text" v-model="groupName" placeholder="Enter group name" />
        </div>
        <button @click="createConversation">Create</button>
        <button @click="cancelSelection">Cancel</button>
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
            if(isMyMessage && msg.read_status && msg.read_status.length > 0) {
              const otherUser = msg.read_status.find(r => r.user_id !== this.currentUserId);
              if(otherUser){
                is_read = otherUser?.is_read === 1 || otherUser?.is_read === true;  
                is_delivered = otherUser?.is_delivered === 1 || otherUser?.is_delivered === true;
              }
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
          content: "Hello!",
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

        this.chats.unshift({
          conversationId: data.ConversationId,
          nameChat: nSelected === 1 ? this.selectedUsers[0] : this.groupName,
          lastMessage: null,
          messages: [],
        });

        this.showUserSelection = false;
        this.selectedUsers = [];
        this.groupName = "";
      } catch (error) {
        this.errorMessage = "Errore durante la creazione della chat: " + error.message;
      }
    },
    cancelSelection() {
      this.showUserSelection = false;
      this.selectedUsers = [];
      this.groupName = "";
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

    }


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
  width: 35px;
  height: 35px;
  border-radius: 50%;
  object-fit: cover;
}

/* Nome utente */
.user-name {
  font-size: 16px;
  font-weight: 500;
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

.message {
  max-width: 70%;
  padding: 10px;
  border-radius: 10px;
  font-size: 14px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
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
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 5px;
  width: 300px;
  text-align: center;
}

.modal-content h3 {
  margin-top: 0;
}

.modal-content ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.modal-content li {
  margin-bottom: 10px;
}

.modal-content button {
  margin-top: 10px;
  padding: 10px 15px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}

.modal-content button:first-of-type {
  background: #007bff;
  color: white;
}

.modal-content button:last-of-type {
  background: #ccc;
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
