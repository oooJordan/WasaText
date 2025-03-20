<template>
	<div class="chat-container">
	  <!-- Sidebar con le chat -->
	  <div class="sidebar">
		<div class="sidebar-header">
		  <h2>Chats</h2>
		  <button @click="startNewChat">➕ New Chat</button>
		</div>
		<ul class="chat-list" v-if="chats.length > 0">
		  <li v-for="chat in chats" :key="chat.id" @click="selectChat(chat.id)" :class="{ active: chat.id === currentChat?.id }">
			<div class="chat-item">
			  <div class="chat-avatar">💬</div>
			  <div class="chat-info">
				<strong>{{ chat.name }}</strong>
				<p v-if="chat.lastMessage">{{ chat.lastMessage.text }}</p>
			  </div>
			</div>
		  </li>
		</ul>
		<p v-else class="no-chats">You have no conversations yet. Start one now!</p>
	  </div>
  
	  <!-- Area principale della chat -->
	  <div class="chat-window" v-if="currentChat">
		<div class="chat-header">
		  <h2>{{ currentChat.name }}</h2>
		</div>
		<div class="messages">
		  <div v-for="message in currentChat.messages" :key="message.id" :class="['message', message.sender === 'You' ? 'sent' : 'received']">
			<p>{{ message.text }}</p>
			<span class="message-time">12:34</span>
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
			  <label class = "user-label">
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
		errorMessage: ""
	  };
	},
	created() {
		this.token = Number(localStorage.getItem("token"));
		if (!this.token || this.token <= 0) {
		  this.$router.push("/login");
		} else {
			this.fetchChats();
		}
	},
	methods: {
	  async fetchChats() {
		try {

		  const token = localStorage.getItem("token");
		  if(!token){
			throw new Error("Unauthorized");
		  }
		  const response = await fetch(`${__API_URL__}/conversations`,{
			headers: {
			  Authorization: `Bearer ${token}`,
		  },

		});
		  this.chats = response.data.conversation;
		} catch (error) {
		  console.error("Error fetching chats:", error);
		}
	  },
	async fetchUsers() {
		try {
			const token = localStorage.getItem("token");
			const response = await this.$axios.get("/users", {
				headers: { Authorization: `Bearer ${token}` },
			});

			// Controllo se la chiave 'users' esiste dentro response.data
			if (!response.data || !response.data.users) {
				throw new Error("La chiave 'users' non esiste nella risposta");
			}

			// Assegno gli utenti
			this.users = response.data.users;
		} catch (error) {
			console.error("Error fetching users:", error);
		}
	},
	  selectChat(chatId) {
		this.currentChat = this.chats.find(chat => chat.id === chatId);
	  },
	  async sendMessage() {
		if (!this.newMessage.trim()) return;
		try {
		  this.currentChat.messages.push({
			sender: "You",
			text: this.newMessage,
			id: Date.now(),
		  });
		  this.newMessage = "";
		} catch (error) {
		  console.error("Error sending message:", error);
		}
	  },
	  startNewChat() {
		this.showUserSelection = true;
		this.fetchUsers();
	  },
	  async createConversation() {
		if (this.selectedUsers.length === 0) {
		  this.errorMessage = "Please select at least one user.";
		  return;
		}
  
		const chatType = this.selectedUsers.length === 1 ? "private_chat" : "group_chat";
		const conversationRequest = {
		  chatType: chatType,
		  usersname: this.selectedUsers,
		  startMessage: {
			media: "text",
			content: "Hello!",
		  },
		};
  
		if (chatType === "group_chat") {
		  if (!this.groupName.trim()) {
			this.errorMessage = "Group name is required for group chat.";
			return;
		  }
		  conversationRequest.groupName = this.groupName;
		  conversationRequest.imageGroup = "https://cdn.raceroster.com/assets/images/team-placeholder.png";
		}
  
		try {
		  const response = await this.$axios.post("/conversations", conversationRequest);
		  this.chats.push({
			id: response.data.ConversationId,
			name: chatType === "private_chat" ? this.selectedUsers[0] : this.groupName,
			messages: [],
		  });
		  this.showUserSelection = false;
		  this.selectedUsers = [];
		  this.groupName = "";
		} catch (error) {
		  if (error.response) {
			switch (error.response.status) {
			  case 400:
				this.errorMessage = "Form error, please check all fields and try again.";
				break;
			  case 401:
				this.errorMessage = "Unauthorized, please log in.";
				break;
			  case 409:
				this.errorMessage = "Conflict, conversation already exists.";
				break;
			  case 500:
				this.errorMessage = "Internal server error, please try again later.";
				break;
			  default:
				this.errorMessage = "An unexpected error occurred.";
			}
		  } else {
			this.errorMessage = "An unexpected error occurred.";
		  }
		}
	  },
	  cancelSelection() {
		this.showUserSelection = false;
		this.selectedUsers = [];
		this.groupName = "";
		this.errorMessage = "";
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
	width: 300px;
	background: #2c3e50;
	color: white;
	display: flex;
	flex-direction: column;
	padding: 10px;
  }
  
  .sidebar-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 10px;
  }
  
  .sidebar-header h2 {
	margin: 0;
	font-size: 1.2em;
  }
  
  .sidebar-header button {
	background: #1abc9c;
	border: none;
	padding: 5px 10px;
	color: white;
	cursor: pointer;
	border-radius: 5px;
  }

  /* Lista utenti nella modale */
.user-list {
  list-style: none;
  padding: 0;
  margin: 10px 0;
  max-height: 300px;
  overflow-y: auto;
}

/* Singolo elemento utente */
.user-item {
  padding: 8px;
  border-radius: 8px;
  transition: background 0.2s ease-in-out;
}

.user-item:hover {
  background: #f1f1f1;
}

/* Stile label */
.user-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  width: 100%;
}

/* Stile checkbox */
.user-checkbox {
  width: 18px;
  height: 18px;
  accent-color: #007bff;
}

/* Avatar utente */
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
	flex: 1;
	display: flex;
	flex-direction: column;
	background: white;
	border-left: 1px solid #ddd;
  }
  
  .chat-header {
	padding: 15px;
	background: #3498db;
	color: white;
	font-size: 1.5em;
  }
  
  .messages {
	flex-grow: 1;
	overflow-y: auto;
	padding: 10px;
	display: flex;
	flex-direction: column;
	gap: 10px;
	margin-bottom: 60px;
  }
  
  .message {
	max-width: 70%;
	padding: 10px;
	border-radius: 10px;
	font-size: 14px;
	display: flex;
	flex-direction: column;
  }
  
  .message.sent {
	align-self: flex-end;
	background: #007bff;
	color: white;
  }
  
  .message.received {
	align-self: flex-start;
	background: #ecf0f1;
	color: black;
  }
  
  .message-time {
	font-size: 10px;
	margin-top: 5px;
	opacity: 0.6;
	text-align: right;
  }
  
  /* Input messaggi */
  .input-area {
	display: flex;
	padding: 10px;
	background: #f8f9fa;
	border-top: 1px solid #ddd;
	width: calc(100% - 300px);
	position: fixed;
	bottom: 0;
	right: 0;
	box-sizing: border-box;
  }
  
  .input-area input {
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

  /* Modal per la selezione degli utenti */
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
  </style>
