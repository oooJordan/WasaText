  <template>
    <div class="chat-container">
      <div class="sidebar">
        <div class="sidebar-header">
          <div class="user-info">
            <!-- Immagine profilo -->
            <img
              :src="profileImage"
              alt="Profile"
              class="profile-image"
              @click="openImageModal"
            />
            <!-- Nome utente -->
            <span class="username-display">{{ currentUser }}</span>
          </div>

          <div class="menu-icon" @click="toggleUserMenu">🛠️</div>

          <div v-if="showUserMenu" class="dropdown-menu user-dropdown">
            <p @click="logout">🚪 Logout</p>
            <p @click="openChangeUsernameModal">🖊️ Cambia nome</p>
            <p @click="openChangeProfileImageModal">🖼️ Cambia immagine profilo</p>
          </div>
        </div>

        <div class="search-bar">
          <div class="search-input-wrapper">
            <span class="search-icon">🔍</span>
            <input
              v-model="searchquery"
              @input="searchUsers"
              type="text"
              placeholder="Search chat..."
              class="search-input"
            />
          </div>
        </div>  

        <ul class="chat-list" v-if="filteredChats.length > 0">
          <li
            v-for="chat in filteredChats"
            :key="chat.conversationId"
            @click="handleChatClick(chat)"
            :class="{ active: chat.conversationId === currentChat?.conversationId }"
          >
            <div class="chat-item">
              <div class="chat-avatar">
                <!-- Immagine profilo di una chat dentro una chat -->
                <img
                  v-if="chat.chatType === 'group_chat' && chat.profileimage"
                  :src="chat.profileimage"
                  alt="Group"
                  class="chat-avatar-img"
                />
                <img
                  v-else-if="chat.chatType === 'private_chat' && chat.profileimage"
                  :src="chat.profileimage"
                  alt="Private"
                  class="chat-avatar-img"
                />

              </div>
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
          <div class="group-header-wrapper">
            <div class="group-name-section" style="display: flex; align-items: center; gap: 10px;">
              <!-- Immagine gruppo -->
              <img
                v-if="currentChat.chatType === 'group_chat' && currentChat.profileimage"
                :src="currentChat.profileimage"
                alt="Group"
                class="group-avatar-img"
                style="width: 40px; height: 40px; border-radius: 50%; object-fit: cover;"
              />

              <!-- Immagine chat privata -->
              <img
                v-if="currentChat.chatType === 'private_chat' && currentChat.profileimage"
                :src="currentChat.profileimage"
                alt="Private Chat"
                class="group-avatar-img"
                style="width: 40px; height: 40px; border-radius: 50%; object-fit: cover;"
              />

              <!-- Nome gruppo -->
              <template v-if="currentChat.chatType === 'group_chat'">
                <template v-if="editingGroupName">
                  <input
                    v-model="editedGroupName"
                    @blur="saveGroupName"
                    @keyup.enter="saveGroupName"
                    class="edit-group-name-input"
                  />
                </template>
                <template v-else>
                  <h2 @click="enableNameEdit" class="editable">
                    {{ currentChat.nameChat }}
                  </h2>
                </template>
              </template>
              <template v-else>
                <h2>{{ currentChat.nameChat }}</h2>
              </template>

            </div>


            <div v-if="currentChat.chatType === 'group_chat'" class="group-menu-wrapper">
              <div class="menu-icon" @click="toggleGroupMenu">⚙️</div>
              <div v-if="showGroupMenu" class="dropdown-menu group-dropdown">
                <p @click="openAddMembersModal">➕ Aggiungi membri</p>
                <p @click="enableNameEdit">🖊️ Modifica nome</p>
                <p @click="openChangeImageModalGroup">🖼️ Cambia immagine</p>
                <p @click="leaveGroup">🚪 Esci dal gruppo</p>
              </div>
            </div>

            <!-- Modale per cambiare l'immagine del gruppo -->
            <div v-if="showChangeImageGroupModal" class="modal">
              <div class="modal-change-image-profile">
                <h3>Cambia immagine del gruppo</h3>
                
                <!-- Input per selezionare una nuova immagine -->
                <input type="file" ref="imageInput" @change="handleProfileImageUpload($event, 'group')" />
                
                <!-- Messaggio di errore -->
                <p v-if="uploadError" class="error-message">{{ uploadError }}</p>
                
                <!-- Bottoni per confermare o annullare -->
                <div class="modal-buttons">
                  <button @click="confirmProfileImageGroupChange">Salva</button>
                  <button @click="showChangeImageGroupModal=false">Annulla</button>
                </div>
              </div>
            </div>
          </div>
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
                <div class="message-content">
                  <!-- Messaggio di testo -->
                  <p v-if="message.media === 'text' || message.media === 'gif_with_text'">{{ sanitizeContent(message.content) }}</p>
                  
                  <!-- PEr vedere se un messaggio è una gif o una gif con testo -->
                  <img
                    v-if="message.media === 'gif' || message.media === 'gif_with_text'"
                    :src="message.image"
                    alt="immagine del messaggio"
                  />
                </div>
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


        <div v-if="selectedGifUrl" class="image-preview">
            <img :src="selectedGifUrl" alt="Anteprima immagine" />
            <button class="remove-image-btn" @click="selectedGifUrl = null">✕</button>
        </div>

        <div class="input-area">
          <div class="input-controls">
            <input
              type="text"
              v-model="newMessage"
              @keyup.enter="sendMessage"
              placeholder="Type a message..."
              class="message-input"
            />

            <label class="file-label-message">
              📎
              <input
                type="file"
                @change="handleProfileImageUpload($event, 'message')"
                style="display: none;"
              />
            </label>

            <button class="send-button" @click="sendMessage">Send</button>
          </div>
        </div>


      </div>


      <div class="chat-window empty" v-else>
        <p>Select a chat to start messaging</p>
      </div>

      <!-- Modal per inoltro messaggio -->
      <div v-if="showForwardModal" class="modal">
        <div class="modal-content-forward horizontal-layout">
          <h3>Inoltra messaggio</h3>

          <div class="forward-columns">
            <!-- Colonna sinistra: Chat esistenti -->
            <div class="forward-column">
              <h4>Chat esistenti</h4>
              <ul class="conversation-list">
                <li
                  v-for="chat in chats"
                  :key="chat.conversationId"
                  :class="{ selected: selectedForwardChatIds.includes(chat.conversationId) }"
                  @click="toggleForwardSelection(chat.conversationId)"
                  class="conversation-item"
                >
                  {{ chat.nameChat }}
                </li>
              </ul>
            </div>

            <!-- Colonna destra: Tutti gli utenti -->
            <div class="forward-column">
              <h4>Altri utenti</h4>
              <ul class="conversation-list">
                <li
                  v-for="user in availableForwardUsers"
                  :key="user.user_id"
                  :class="{ selected: selectedForwardUsernames.includes(user.nickname) }"
                  @click="toggleForwardUserSelection(user.nickname)"
                  class="conversation-item"
                >
                  {{ user.nickname }}
                </li>
              </ul>
            </div>
          </div>

          <!-- Bottoni -->
          <div class="modal-buttons">
            <button @click="confirmForward">Inoltra</button>
            <button @click="cancelForward">Annulla</button>
          </div>
        </div>
      </div>


      <!-- Modal per la creazione di una nuova chat -->
      <div v-if="showUserSelection" class="modal">
        <div class="modal-content">
          <!-- Colonna di sinistra -->
          <div class="left-column">
            <h3>Crea nuova chat</h3>

            <!-- Per selezionare il tipo chat -->
            <div class="chatType-buttons">
              <button
                :class="{ active: chatType === 'private_chat' }"
                @click="chatType = 'private_chat'"
              >
                Chat Privata
              </button>
              <button
                :class="{ active: chatType === 'group_chat' }"
                @click="chatType = 'group_chat'"
              >
                Gruppo
              </button>
            </div>

            <!-- Barra di ricerca utenti -->
            <input
              v-model="searchnome"
              placeholder="Cerca utente..."
              class="search-input-wrapper"
            />

            <!-- Lista utenti filtrata -->
            <ul class="conversation-list">
              <li
                v-for="user in filteredUserList"
                :key="user.user_id"
                :class="['conversation-item', selectedUsers.includes(user.nickname) ? 'selected' : '']"
                @click="toggleUserSelection(user.nickname)"
              >
                {{ user.nickname }}
              </li>
            </ul>
          </div>

          <!-- Colonna destra -->
          <div class="right-column">
            <!-- Nome e immagine gruppo  -->
            <div v-if="chatType === 'group_chat'" class="group-fields">
              <input
                v-model="groupName"
                type="text"
                placeholder="Nome del gruppo"
                class="search-input-wrapper"
              />
              <!-- Upload immagine gruppo -->
              <div class="upload-section">
                  <label class="file-label">
                    📁 Carica immagine gruppo
                    <input
                      type="file"
                      @change="handleProfileImageUpload($event, 'group')"
                      accept="image/*"
                      style="display: none;"
                    />
                  </label>

                  <!-- Messaggio di errore se c'è stato un errore nell'upload -->
                  <p v-if="uploadError" class="error-message">{{ uploadError }}</p>

                  <!-- Messaggio di conferma immagine caricata -->
                  <p v-if="selectedImageGroupNewChat" class="preview-label">Immagine caricata!</p>
                </div>
            </div>

            <!-- Messaggio iniziale -->
            <textarea
              v-model="startMessageText"
              placeholder="Scrivi un messaggio iniziale..."
              class="search-input-wrapper"
              rows="4"
            ></textarea>

            <!-- Upload immagine per il messaggio iniziale -->
            <div class="upload-section">
              <label class="file-label">
                🖼️ Aggiungi immagine al messaggio iniziale
                <input
                  type="file"
                  @change="handleProfileImageUpload($event, 'initialMessageImage')"
                  accept="image/*"
                  style="display: none;"
                />
              </label>

              <!-- Messaggio di errore -->
              <p v-if="uploadError" class="error-message">{{ uploadError }}</p>

            </div>

            <!-- Anteprima immagine messaggio iniziale -->
            <div v-if="selectedImageGroupNewChat" class="image-preview-upload">
              <img :src="selectedImageGroupNewChat" alt="Anteprima immagine" />
              <button class="remove-image-btn" @click="selectedImageGroupNewChat = null">✕</button>
            </div>

            <!-- Messaggi di errore -->
            <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>

            <!-- Bottoni per creazione o annulla -->
            <div class="modal-buttons">
              <button class="primary-btn" @click="createConversation">Crea</button>
              <button class="secondary-btn" @click="cancelSelection">Annulla</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Modal per aggiungere membri a un gruppo esistente -->
      <div v-if="showAddMembersModal" class="modal">
        <div class="modal-content-forward">
          <h3>Aggiungi membri al gruppo</h3>

          <h4>Utenti</h4>
            <ul class="conversation-list">
              <li
                v-for="user in allUsers"
                :key="user.user_id"
                :class="{ selected: selectedForwardUsernames.includes(user.nickname) }"
                @click="toggleForwardUserSelection(user.nickname)"
                class="conversation-item"
              >
                {{ user.nickname }}
              </li>
            </ul>

          <div class="modal-buttons">
            <button class="primary-btn" @click="confirmAddMembers">Aggiungi</button>
            <button class="secondary-btn" @click="cancelAddMembers">Annulla</button>
          </div>
        </div>
      </div>
      <!-- Modal per cambiare il nome utente -->
      <div v-if="showChangeUsernameModal" class="modal">
        <div class="modal-change-username">
          <h3>Cambia il mio nome utente</h3>

          <input
            v-model="newUsername"
            type="text"
            placeholder="Inserisci il nuovo nome"
            class="styled-input"
          />

          <!-- Messaggio di errore in caso di nome utente già esistente -->
          <p v-if="usernameError" class="error-message">{{ usernameError }}</p>

          <div class="modal-buttons">
            <button class="primary-btn" @click="updateUsername">Salva</button>
            <button class="secondary-btn" @click="showChangeUsernameModal = false">Annulla</button>
          </div>
        </div>

      </div>
      <!-- Modal per immagine profilo ingrandita -->
      <div v-if="showImageModal" class="modal-image">
        <div class="modal-content-image">
          <img :src="profileImage" class="modal-profile-image" />
          <button class="close-btn" @click="showImageModal = false">Chiudi</button>
        </div>
      </div>

      <!-- Modale per cambiare immagine profilo -->
      <div v-if="showChangeProfileImageModal" class="modal">
        <div class="modal-change-image-profile">
          <h3>Carica una nuova immagine profilo</h3>

          <input type="file" @change="handleProfileImageUpload($event, 'profile')" />

          <p v-if="uploadError" class="error-message">{{ uploadError }}</p>

          <div class="modal-buttons">
            <button @click="confirmProfileImageChange">Salva</button>
            <button @click="showChangeProfileImageModal = false">Annulla</button>
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
      allUsers: [],
      selectedUsers: [],
      groupName: "",
      errorMessage: "",
      searchquery: "",
      messageArray: [],
      selectedMessageOptions: null,
      hoveredMessage: null,
      reactionMessageId: null,
      emojiOptions: ["👍", "😂​", "❤️​"],
      showGroupMenu: false,
      editingGroupName: false,
      editedGroupName: '',
      showAddMembersModal: false,
      showForwardModal: false,
      messageToForward: null,
      selectedForwardUsernames: [],
      selectedForwardChatIds: [],
      chatType: "",
      searchnome: "",
      showUserMenu: false, 
      newUsername: "",
      showChangeProfileImageModal: false,
      selectedProfileImage: null,
      uploadError: "",
      currentUser:"",
      profileImage: "",
      showImageModal: false,
      showChangeImageGroupModal: false,
      selectedImageGroup: null,
      selectedGifUrl: null,
      newImageFile: null,
      showChangeUsernameModal: false,
      usernameError: "",
      startMessageText: "",
      selectedImageGroupNewChat: null,
    };
  },
  created() {
    this.token = Number(localStorage.getItem("token"));
    this.currentUser = localStorage.getItem("username") || "You";
    if (!this.token || this.token <= 0) {
      this.$router.push("/login");
    } else {
      this.fetchChats();
      this.fetchAllUsers(); 
    }
  },
  computed:
  {
    filteredChats() {
      if (!Array.isArray(this.chats)) return [];
      return this.chats.filter(chat =>
        chat.nameChat?.toLowerCase().includes(this.searchquery.toLowerCase())
      );
    },
    filteredUserList() {
      return this.users.filter(user => user.nickname.toLowerCase().includes(this.searchnome.toLowerCase())
      )},
    availableForwardUsers() {
      const existingPrivateUsernames = this.chats.filter(chat => chat.ChatType === "private_chat").map(chat => chat.nameChat);

      return this.allUsers.filter( user => user.nickname !== this.currentUser && !existingPrivateUsernames.includes(user.nickname));
    }
  },
  methods: {
    async fetchMessageHistory(conversation_id) {
      try {
        const token = localStorage.getItem("token");
        if (!token) throw new Error("Non autorizzato");

        const response = await this.$axios.get(`/conversation/${conversation_id}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        const data = response.data;

        const chatFromSidebar = this.chats.find(c => c.conversationId === conversation_id);
        if (!chatFromSidebar) {
          console.warn("Chat non trovata nella sidebar");
        }

        this.currentChat = {
          ...chatFromSidebar,
          chatType: chatFromSidebar?.chatType?.chatType || chatFromSidebar?.chatType || "private_chat",
          messages: data.messages.map(msg => {
            const isMyMessage = msg.username === this.currentUser;
            let is_read = false;
            let is_delivered = false;

            if (isMyMessage && msg.read_status?.length > 0) {
              const otherUsers = msg.read_status.filter(r => r.user_id !== this.currentUserId);
              is_read = otherUsers.length > 0 && otherUsers.every(r => r.is_read);
              is_delivered = otherUsers.length > 0 && otherUsers.every(r => r.is_delivered);
            }

            // Gestione del content basata sul tipo di media
            const trimmedContent = (msg.content || "").trim();
            switch (msg.media) {
              case "text":
                if (!trimmedContent) {
                  msg.content = "[Messaggio vuoto]";
                }
                break;

              case "gif":
                if (!trimmedContent) {
                  msg.content = "[Foto]";
                }
                break;

              case "gif_with_text":
                msg.content = trimmedContent ? `[Foto] ${trimmedContent}` : "[Foto]";
                break;

              default:
                console.warn("Media sconosciuto:", msg.media);
                break;
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
        if (!token) throw new Error("Non autorizzato");

        const response = await this.$axios.get(`/conversations`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        const data = response.data;

        this.chats = Array.isArray(data.conversation) ? data.conversation.map(chat => {
          const lastMsg = chat.lastMessage || {};
          const trimmedContent = (lastMsg.content || "").trim();

          switch (lastMsg.media) {
            case "gif":
              lastMsg.content = trimmedContent || "[Foto]";
              break;

            case "gif_with_text":
              lastMsg.content = trimmedContent ? `[Foto] ${trimmedContent}` : "[Foto]";
              break;

            case "text":
              if (!trimmedContent) {
                lastMsg.content = "[Messaggio vuoto]";
              }
              break;

            default:
              break;
          }

          return {
            ...chat,
            lastMessage: lastMsg,
            chatType: chat.chatType?.chatType || chat.chatType || chat.ChatType || "private_chat"
          };
        }) : [];

      } catch (error) {
        console.error("Errore nel fetch delle chat:", error);
      }
    },
    async fetchUsers({ isForNewChat = false } = {}) {
      try {
        const token = localStorage.getItem("token");

        const response = await this.$axios.get("/users", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        const data = response.data;

        if (!data || !data.users) {
          throw new Error("La chiave 'users' non esiste nella risposta");
        }

        let filteredUsers = data.users.filter(u => u.nickname !== this.currentUser);

        if (!isForNewChat && this.currentChat?.chatType === 'group_chat') {
          const existingUsernames = this.currentChat.users.map(u => u.nickname);
          filteredUsers = filteredUsers.filter(u => !existingUsernames.includes(u.nickname));
        }

        this.users = filteredUsers;

      } catch (error) {
        console.error("Errore nel fetch degli utenti:", error);
      }
    },
    selectChat(chat) {
      this.currentChat = chat
      if (chat.conversationId) {
        this.fetchMessageHistory(chat.conversationId);
      } else {
        console.warn("Nessuna chat trovata con id:", chat.conversationId);
      }
    },
    async sendMessage() {
      if (!this.newMessage.trim() && !this.selectedGifUrl) return;

      let mediaType = "text";
      if (this.selectedGifUrl && this.newMessage.trim()) {
        mediaType = "gif_with_text";
      } else if (this.selectedGifUrl && !this.newMessage.trim()) {
        mediaType = "gif";
      }

      try {
        const token = localStorage.getItem("token");
        if (!token) throw new Error("Non autorizzato");

        if (!this.currentChat?.conversationId) {
          return;
        }

        const messagePayload = {
          content: this.newMessage.trim(),
          media: mediaType,
          image: this.selectedGifUrl || ""
        };

        const response = await this.$axios.post(`/conversation/${this.currentChat.conversationId}`,
          messagePayload,
          {
            headers: {
              Authorization: `Bearer ${token}`
            }
          }
        );

        const data = response.data;

        this.currentChat.messages.push({
          message_id: data.messageId || Date.now(),
          username: this.currentUser,
          content: this.newMessage.trim(),
          media: mediaType,
          image: this.selectedGifUrl || "",
          timestamp: new Date().toISOString(),
        });

        const chat = this.chats.find(c => c.conversationId === this.currentChat.conversationId);
        if (chat) {
          const preview = this.getPreview({
            media: mediaType,
            content: this.newMessage.trim()
          });

          chat.lastMessage = {
            content: preview,
            timestamp: new Date().toISOString()
          };

          const index = this.chats.findIndex(c => c.conversationId === chat.conversationId);
          if (index > -1) {
            const [updateChat] = this.chats.splice(index, 1);
            this.chats.unshift(updateChat);
          }
        }

        this.newMessage = "";
        this.selectedGifUrl = null;
        this.newImageFile = null;
        if (this.$refs.imageInput) {
          this.$refs.imageInput.value = null;
        }

        this.$nextTick(() => this.scrollToLastMessageWithSmooth());
        await this.fetchChats();


      } catch (error) {
        console.error("Errore nell'invio del messaggio:", error);
      }
    },
    startNewChat() {
      this.showUserSelection = true;
      this.fetchUsers({ isForNewChat: true });
    },
    async createConversation() {
      this.errorMessage = "";

      if (this.selectedUsers.length === 0) {
        this.errorMessage = "Seleziona almeno un utente";
        return;
      }
      if (!this.startMessageText.trim()  && !this.selectedImageGroupNewChat) {
        this.errorMessage = "Devi scrivere un messaggio iniziale";
        return;
      }

      if (this.chatType === "private_chat" && this.selectedUsers.length !== 1) {
        this.errorMessage = "La chat privata deve avere esattamente un partecipante";
        return;
      }
      if (this.chatType === "group_chat" && !this.groupName.trim()) {
        this.errorMessage = "Devi inserire un nome al gruppo";
        return;
      }

      let mediaType = "text";
      if (this.selectedImageGroup && this.startMessageText.trim()) {
        mediaType = "gif_with_text";
      } else if (this.selectedImageGroup && !this.startMessageText.trim()) {
        mediaType = "gif";
      }

      const conversationRequest = {
        chatType: { ChatType: this.chatType },
        groupName: this.chatType === "group_chat" ? this.groupName : "",
        imageGroup: this.chatType === "group_chat" ? (this.selectedImageGroup || "") : "",
        usersname: this.selectedUsers,
        startMessage: {
          media: this.selectedImageGroupNewChat && this.startMessageText.trim()
            ? "gif_with_text"
            : this.selectedImageGroupNewChat
            ? "gif"
            : "text",
          content: this.startMessageText.trim(),
          image: this.selectedImageGroupNewChat || ""
        }

      };
      try {
        const token = localStorage.getItem("token");
        if (!token) throw new Error("Non sei autorizzato");

        const response = await this.$axios.post(`/conversations`, conversationRequest, {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });

        const data = response.data;

        if (!data || !data.ConversationId) {
          throw new Error("La risposta API non contiene un ConversationId valido");
        }

        this.showUserSelection = false;
        this.selectedUsers = [];
        this.groupName = "";

        await this.fetchChats();
        const newChat = this.chats.find(c => c.conversationId === data.ConversationId);
        if (newChat) {
          this.selectChat(newChat);
        }

      } catch (error) {
        this.errorMessage = "Errore durante la creazione della chat: " + error.message;
      }
    },
    async deleteMessage(message) {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }

        await this.$axios.delete(`/conversation/${this.currentChat.conversationId}/messages/${message.message_id}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        this.currentChat.messages = this.currentChat.messages.filter(m => m.message_id !== message.message_id);

        if (this.currentChat.messages.length === 0) {
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
    async addReaction(message, emoji) {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }

        const exist = message.comments?.find(c => c.username === this.currentUser);
        if (exist && exist.emojiCode === emoji) {
          return this.removeReaction(message);
        }

        await this.$axios.put(
          `/conversation/${this.currentChat.conversationId}/messages/${message.message_id}/comment`,
          { emojiCode: emoji },
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );

        if (!message.comments) {
          message.comments = [];
        }

        if (exist) {
          exist.emojiCode = emoji;
        } else {
          message.comments.push({ username: this.currentUser, emojiCode: emoji });
        }

        this.reactionMessageId = null;
        this.selectedMessageOptions = null;

      } catch (err) {
        console.error("Errore durante l'aggiunta dell'emoji: ", err);
      }
    },
    async removeReaction(message) {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }

        await this.$axios.delete(`/conversation/${this.currentChat.conversationId}/messages/${message.message_id}/comment`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (message.comments) {
          message.comments = message.comments.filter(c => c.username !== this.currentUser);
        }

        this.reactionMessageId = null;
        this.selectedMessageOptions = null;

      } catch (err) {
        console.error("Errore durante la rimozione dell'emoji: ", err);
      }
    },
    async leaveGroup() {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }

        await this.$axios.delete(`/conversation/${this.currentChat.conversationId}/membership`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        this.chats = this.chats.filter(chat => chat.conversationId !== this.currentChat.conversationId);
        if (this.currentChat.conversationId) {
          this.currentChat = null;
        }

      } catch (err) {
        console.error("Errore durante l'uscita dal gruppo:", err);
      }
    },
    async updateGroupName(conversationId, newName) {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }

        await this.$axios.put(`/conversation/${this.currentChat.conversationId}/name`, { newUsername: newName },
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );

        const chat = this.chats.find(c => c.conversationId === conversationId);
        if (chat) {
          chat.nameChat = newName;
        }

        if (this.currentChat?.conversationId === conversationId) {
          this.currentChat.nameChat = newName;
        }
        const chatInSidebar = this.chats.find(c => c.conversationId === conversationId);
          if (chatInSidebar) {
            chatInSidebar.nameChat = newName;
          }


      } catch (err) {
        console.error("Errore durante l'aggiornamento del nome del gruppo:", err);
      }
    },
    async addUserToGroup(conversationId, usernameUser) {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          throw new Error("Non sei autorizzato");
        }

        await this.$axios.put(
          `/conversation/${conversationId}/names`,
          { name: usernameUser },
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );

      } catch (err) {
        console.error("Errore durante l'aggiunta di un utente al gruppo: ", err);
      }
    },
    async confirmForward() {
      const token = localStorage.getItem("token");
      const message = this.messageToForward;
      if (!message || (!this.selectedForwardChatIds.length && !this.selectedForwardUsernames.length)) return;

      // Inoltro a chat esistenti
      for (const conversationId of this.selectedForwardChatIds) {
        try {
          await this.$axios.post(`/conversation/${conversationId}/messages/${message.message_id}`, {},
            { headers: { Authorization: `Bearer ${token}` } }
          );
        } catch (err) {
          console.error("Errore inoltro a chat ID:", conversationId, err);
        }
      }

      const createdChats = new Map(); // username → conversationId

      // Creo nuove chat private e salvo il loro ID
      for (const username of this.selectedForwardUsernames) {
        try {
          const response = await this.$axios.post(`/conversations`,
            {
              chatType: { ChatType: "private_chat" },
              groupName: "",
              imageGroup: "",
              usersname: [username],
              startMessage: {
                media: message.media,
                content: this.sanitizeContent(message.content) || "",
                image: message.image || ""
              }
            },
            { headers: { Authorization: `Bearer ${token}` } }
          );


          const conversationId = response.data.ConversationId;
          if (conversationId) {
            createdChats.set(username, conversationId);
          } else {
            console.error("conversationId non presente nella risposta per", username);
          }

        } catch (err) {
          console.error("Errore nella creazione della chat per utente:", username, err);
        }
      }

      for (const username of this.selectedForwardUsernames) {
        if (createdChats.has(username)) {
          // Messaggio già incluso come startMessage, non inoltro di nuovo
          continue;
        }

        const chat = this.chats.find(chat =>
          chat.chatType === "private_chat" &&
          chat.users &&
          chat.users.some(u => (u.nickname || u.username) === username)
        );
        if (chat && chat.conversationId) {
          try {
            await this.$axios.post(`/conversation/${chat.conversationId}/messages/${message.message_id}`, {},
              { headers: { Authorization: `Bearer ${token}` } }
            );
          } catch (err) {
            console.error("Errore inoltro fallback per", username, err);
          }
        }
      }

      await this.fetchChats(); // ricarico per aggiornare le preview

      // Costruzione della preview
      const preview = this.getPreview(message);

      // Aggiorno le preview nelle chat esistenti
      this.selectedForwardChatIds.forEach(conversationId => {
        const chat = this.chats.find(c => c.conversationId === conversationId);
        if (chat) {
          chat.lastMessage = { content: preview, timestamp: new Date().toISOString() };
          const index = this.chats.findIndex(c => c.conversationId === conversationId);
          if (index > -1) {
            const [updateChat] = this.chats.splice(index, 1);
            this.chats.unshift(updateChat);
          }
        }
      });

      // Aggiorno anche le nuove chat
      createdChats.forEach((conversationId, username) => {
        const chat = this.chats.find(c => c.conversationId === conversationId);
        if (chat) {
          chat.lastMessage = { content: preview, timestamp: new Date().toISOString() };
          const index = this.chats.findIndex(c => c.conversationId === conversationId);
          if (index > -1) {
            const [updateChat] = this.chats.splice(index, 1);
            this.chats.unshift(updateChat);
          }
        }
      });

      this.showForwardModal = false;
      this.messageToForward = null;
      this.selectedForwardChatIds = [];
      this.selectedForwardUsernames = [];
    },
    async fetchAllUsers() {
      try {
        const token = localStorage.getItem("token");
        const res = await this.$axios.get("/users", {
          headers: { Authorization: `Bearer ${token}` }
        });

        this.allUsers = res.data.users.filter(u => u.nickname !== this.currentUser);
      } catch (err) {
        console.error("Errore nel fetch utenti:", err);
      }
    },
    async searchUsers() {
      try {
        const token = localStorage.getItem("token");
        if (!token) throw new Error("Non autorizzato");

        if (!this.searchquery.trim()) {
          this.users = [];
          return;
        }

        const response = await this.$axios.get(`/users?name=${encodeURIComponent(this.searchquery)}`,
          {
            headers: {
              Authorization: `Bearer ${token}`
            }
          }
        );

        this.users = response.data.users || [];

      } catch (err) {
        console.error("Errore nella ricerca utenti:", err);
      }
    },
    async updateUsername() {
      this.usernameError = "";

      if (!this.newUsername.trim()) {
        this.usernameError = "Il nome non può essere vuoto.";
        return;
      }

      try {
        const token = localStorage.getItem("token");
        if (!token) throw new Error("Non autorizzato");

        await this.$axios.put( "/username", { newUsername: this.newUsername.trim() },
          {
            headers: {
              Authorization: `Bearer ${token}`
            }
          }
        );

        localStorage.setItem("username", this.newUsername.trim());
        this.currentUser = this.newUsername.trim();
        this.showChangeUsernameModal = false;

      } catch (error) {
        console.error("Errore durante l'aggiornamento del nome utente:", error);
        if (error.response && error.response.status === 409) {
          this.usernameError = "Questo nome utente è già in uso.";
        } else {
          this.usernameError = "Errore durante l'aggiornamento del nome.";
        }
      }
    },
    async handleProfileImageUpload(event, type) {
      const file = event.target.files[0];
      if (!file) return;

      const formData = new FormData();
      formData.append("file", file);

      try {
        const token = localStorage.getItem("token");
        const response = await this.$axios.post("/upload", formData, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        const data = response.data;

        if (type === 'profile') {
          this.selectedProfileImage = data.imageUrl;
        } else if (type === 'group') {
          this.selectedImageGroup = data.imageUrl;
        } else if (type === 'message') {
          this.selectedGifUrl = data.imageUrl;
        } else if (type === 'initialMessageImage') {
          this.selectedImageGroupNewChat = data.imageUrl;
        }

      } catch (err) {
        console.error("Errore upload immagine:", err);
        this.uploadError = err.message || "Errore durante l'upload";
      }finally{
        event.target.value = null;
      }
    },
    async fetchProfileImage() {
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          console.error("Token mancante");
          return;
        }

        const response = await this.$axios.get("/profile_image", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        this.profileImage = response.data.actualImage;

      } catch (error) {
        console.error("Errore nel caricamento dell'immagine profilo:", error);
      }
    },
    async confirmProfileImageChange() {
      if (!this.selectedProfileImage) {
        this.uploadError = "Devi prima selezionare un'immagine!";
        return;
      }

      try {
        const token = localStorage.getItem("token");

        const response = await this.$axios.put(
          "/profile_image",
          { image: this.selectedProfileImage },
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );

        this.fetchProfileImage();
        this.showChangeProfileImageModal = false;
        this.selectedProfileImage = null;

      } catch (err) {
        this.uploadError = "Errore durante aggiornamento immagine profilo";
        console.error(err);
      }
    },
    async confirmProfileImageGroupChange() {
      if (!this.selectedImageGroup) {
        this.uploadError = "Devi prima selezionare un'immagine!";
        return;
      }

      try {
        const token = localStorage.getItem("token");

        const response = await this.$axios.put(`/conversation/${this.currentChat.conversationId}/groupimage`,
          { Image: this.selectedImageGroup },
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );

        this.currentChat.profileimage = this.selectedImageGroup;
        this.fetchChats();
        this.showChangeImageGroupModal = false;
        const chatInSidebar = this.chats.find(c => c.conversationId === this.currentChat.conversationId);
        if (chatInSidebar) {
          chatInSidebar.profileimage = this.selectedImageGroup;
        }

      } catch (err) {
        this.uploadError = "Errore durante aggiornamento immagine profilo";
        console.error(err);
      }
    },
    openImageModal() {
      this.showImageModal = true;
    },
    cancelAddMembers() {
      this.showAddMembersModal = false;
      this.selectedUsers = [];
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
      const maxChars = 25;
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
    toggleOptionsMenu(id) {
      if (this.selectedMessageOptions === id) {
        this.selectedMessageOptions = null;

        if (this.reactionMessageId === id) {
          this.reactionMessageId = null;
        }
      } else {
        this.selectedMessageOptions = id;
      }
    },
    handleChatClick(chat) {
      if (!chat.conversationId) {
        this.currentChat = null;
        return;
      }

      if (this.showForwardModal && this.messageToForward) {
        this.confirmForward(chat.conversationId);
      } else {
        this.selectChat(chat);
      }
    },
    forwardMessage(message) {
      this.messageToForward = message;
      this.showForwardModal = true;
      this.selectedMessageOptions = null;
    },
    cancelForward() {
      this.showForwardModal = false;
      this.messageToForward = null;
      this.selectedForwardChatIds = [];
    },
    toggleForwardSelection(conversationId) {
      const index = this.selectedForwardChatIds.indexOf(conversationId);
      if (index === -1) {
        this.selectedForwardChatIds.push(conversationId);
      } else {
        this.selectedForwardChatIds.splice(index, 1);
      }
    },
    showEmoji(message) {
      this.reactionMessageId = message.message_id;
    },
    toggleGroupMenu() {
      this.showGroupMenu = !this.showGroupMenu;
    },
    toggleUserSelection(nickname) {
      const index = this.selectedUsers.indexOf(nickname);
      if (index > -1) {
        this.selectedUsers.splice(index, 1);
      } else {
        this.selectedUsers.push(nickname);
      }
    },
    toggleForwardUserSelection(nickname) {
      const index = this.selectedForwardUsernames.indexOf(nickname);
      if (index === -1) {
        this.selectedForwardUsernames.push(nickname);
      } else {
        this.selectedForwardUsernames.splice(index, 1);
      }
    },
    enableNameEdit() {
      this.editedGroupName = this.currentChat.nameChat;
      this.editingGroupName = true;
      this.showGroupMenu = false;
    },
    saveGroupName() {
      if (this.editedGroupName && this.editedGroupName !== this.currentChat.nameChat) {
        this.updateGroupName(this.currentChat.conversationId, this.editedGroupName);
      }
      this.editingGroupName = false;
    },
    openAddMembersModal() {
      this.showGroupMenu = false;
      this.showAddMembersModal = true;
      this.fetchUsers();
    },
    openChangeImageModal() {
      this.showChangeImageModal = true;
    },
    openChangeImageModalGroup(){
      this.showChangeImageGroupModal = true;
      this.showGroupMenu = false;
    },
    confirmAddMembers(){
      for (const nickname of this.selectedUsers) {
        this.addUserToGroup(this.currentChat.conversationId, nickname);
      }
      this.showAddMembersModal = false;
      this.selectedUsers = [];

      this.fetchMessageHistory(this.currentChat.conversationId);
    },
    toggleUserMenu() {
    this.showUserMenu = !this.showUserMenu;
    },
    openChangeUsernameModal() {
      this.newUsername = "";
    this.showChangeUsernameModal = true;
    this.showUserMenu = false;
    },
    openChangeProfileImageModal() {
      this.showChangeProfileImageModal = true;
      this.uploadError = "";
      this.selectedProfileImage = null;
      this.showUserMenu = false;
    },
    getPreview(msg) {
      const content = msg.content.trim();
      if (msg.media === "gif") {
        // Se non c'è testo, metto [Foto]
        // Se c’è già "[Foto]" in content, non lo ri-aggiungo
        if (!content) return "[Foto]";
        if (content.startsWith("[Foto]")) return content; 
        return "[Foto] " + content;
      }
      if (msg.media === "gif_with_text") {
        if (content.startsWith("[Foto]")) return content;
        return "[Foto] " + content;
      }
      return content;
    },
    sanitizeContent(rawContent) {
      return rawContent.replace(/^(\[Foto\]\s*)+/, "").trim();
    }
    

  },

  mounted() {
    this.fetchChats();
    this.fetchProfileImage();
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

.modal-image {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.modal-content-image {
  background: white;
  padding: 1rem;
  border-radius: 12px;
  text-align: center;
}

.modal-profile-image {
  max-width: 90vw;
  max-height: 80vh;
  border-radius: 8px;
}

.close-btn {
  margin-top: 1rem;
  padding: 0.5rem 1rem;
  background: #333;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}


.modal-buttons {
  display: flex;
  justify-content: space-between;
  gap: 10px;
}

.modal-buttons button {
  flex: 1;
  padding: 10px;
  border: none;
  border-radius: 6px;
  font-weight: bold;
  cursor: pointer;
  transition: background 0.3s;
}

.modal-buttons button:first-of-type {
  background: #007bff;
  color: white;
}

.modal-buttons button:last-of-type {
  background: #ccc;
  color: black;
}

.modal-buttons button:hover {
  opacity: 0.9;
}

.chat-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.conversation-list {
  list-style: none;
  padding: 0;
  margin-top: 1rem;
  max-height: 300px;
  overflow-y: auto;
}

.conversation-item {
  background: #f1f1f1;
  padding: 12px 16px;
  margin-bottom: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s ease;
}

.conversation-item:hover {
  background: #e0e0e0;
}

.conversation-item.selected {
  background: #007bff;
  color: white;
}


.modal-buttons button {
  margin-top: 10px;
  padding: 8px 14px;
}

.chat-list ul{
  overflow-y: auto;
  overflow-x: none;
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

.user-info {
  background: rgba(43, 218, 194, 0.638); 
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 12px;
  padding: 6px 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.25s ease;
  border: 1px solid rgba(255, 255, 255, 0.15);
  position: relative;
}

.user-info:hover {
  transform: scale(1.02);
  box-shadow: 0 6px 18px rgba(45, 212, 191, 0.4);
  background: rgba(49, 243, 217, 0.35); 
}

.user-info .profile-image {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1.5px solid white;
  object-fit: cover;
  transition: transform 0.2s ease;
}

.user-info .profile-image:hover {
  transform: scale(1.1);
}

.username-display {
  font-weight: 600;
  font-size: 1rem;
  color: #ffffff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.profile-image {
  width: 40px;
  height: 40px;
  object-fit: cover;
  border-radius: 50%;
  border: 2px solid #ccc;
}

.username-display {
  font-weight: bold;
  color: #fff;
}

.chat-avatar-img {
  width: 40px;
  height: 40px;
  object-fit: cover;
  border-radius: 50%;
}
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
  padding: 10px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  z-index: 1;
  height: 60px;
}

/* Container per nome gruppo e menu */
.group-header-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

/* Nome gruppo e input inline */
.group-name-section h2 {
  margin: 0;
  font-size: 1.4rem;
  cursor: default;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.group-name-section h2.editable {
  cursor: pointer;
  transition: color 0.2s;
}

.group-name-section h2.editable:hover {
  color: #ecf0f1;
}

.edit-group-name-input {
  font-size: 1.4rem;
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid #ccc;
  max-width: 300px;
}

/* Icona ⋮ e menu */
.group-menu-wrapper {
  position: relative;
  margin-left: 10px;
}

.menu-icon {
  cursor: pointer;
  font-size: 1.5rem;
  padding: 4px 10px;
  user-select: none;
}

.group-dropdown {
  position: absolute;
  top: 30px;
  right: 0;
  background: white;
  color: #333;
  border: 1px solid #ccc;
  border-radius: 6px;
  padding: 5px 0;
  z-index: 100;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
}

.group-dropdown p {
  margin: 0;
  padding: 8px 15px;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.2s;
}

.group-dropdown p:hover {
  background-color: #f2f2f2;
}

.selected {
  background-color: #f0f0f0;
  font-weight: bold;
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
  top: 90%;
  background: white;
  border: 1px solid #ccc;
  box-shadow: 0 2px 6px rgba(0,0,0,0.2);
  z-index: 999;
  border-radius: 20%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.user-dropdown{
  position: absolute;
  right: 0;
  top: 8%;
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

.modal-change-username {
  background: #ffffff;
  border-radius: 16px;
  padding: 24px;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 16px;
  animation: fadeInUp 0.25s ease-out;
}

.modal-change-username h3 {
  font-size: 24px;
  font-weight: bold;
  color: #2c3e50;
  margin: 0;
  text-align: center;
}

.modal-change-username input.styled-input {
  padding: 12px 16px;
  font-size: 1rem;
  border-radius: 10px;
  border: 1px solid #ccc;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  width: 100%;
}

.modal-change-username input.styled-input:focus {
  border-color:#1abc9c;
  box-shadow: 0 0 0 3px rgba(26, 188, 156, 0.3);
  outline: none;
}

.modal-change-username .modal-buttons {
  display: flex;
  justify-content: space-between;
  gap: 10px;
}

.modal-change-username .modal-buttons button {
  flex: 1;
  padding: 12px 0;
  font-size: 1rem;
  font-weight: 600;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.modal-change-username .modal-buttons .primary-btn {
  background-color: #007bff;
  color: white;
}

.modal-change-username .modal-buttons .secondary-btn {
  background-color: #ecf0f1;
  color: #2c3e50;
}

.modal-content-forward {
  background: white;
  border-radius: 10px;
  padding: 20px;
  width: 90%;
  max-width: 600px;
  max-height: 1900px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.2);
}


.forward-columns {
  display: flex;
  gap: 20px;
  margin-top: 10px;
}

.forward-column {
  flex: 1;
  max-height: 700px;
  overflow-y: auto;
  overflow: hidden;
}

.horizontal-layout {
  width: 90%;
  max-width: 900px;
}


.modal-content {
  background: white;
  border-radius: 12px;
  padding: 20px;
  width: 90%;
  max-width: 600px;
  display: flex;
  flex-direction: row;
  gap: 1rem;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.2);
}

.modal-change-image-profile {
  background: white;
  border-radius: 12px;
  padding: 20px;
  width: 90%;
  max-width: 600px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.2);
}

.modal-content h3 {
  margin: 0 0 10px;
  font-size: 22px;
  font-weight: bold;
  text-align: center;
}

.left-column,
.right-column {
  flex: 1;              
  display: flex;
  flex-direction: column;
}

.left-column {
  border-right: 1px solid #ccc;
  padding-right: 1rem;
}

.right-column {
  padding-left: 1rem;
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
  font-weight: bold;
  margin-top: 10px;
}


.search-bar {
  margin: 0 0 15px 0;
}

.search-input-wrapper {
  margin-top: 1rem;
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

.search-results-box {
  background-color: #1abc9c;
  border-radius: 10px;
  padding: 12px;
  margin: 10px 0;
  color: rgb(0, 0, 0);
}

.search-results-title {
  font-weight: bold;
  margin-bottom: 10px;
  font-size: 20px;
}

.chatType-buttons {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

/* Stile dei pulsanti per scegliere chatType */
.chatType-buttons button {
  background-color: #eee;
  color: #333;
  border: 1px solid #ccc;
  border-radius: 4px;
  padding: 0.5rem 1rem;
  cursor: pointer;
  font-weight: bold;
  transition: background-color 0.3s ease, color 0.3s ease;

}

.chatType-buttons button:hover {
  background-color: #ddd;
}

/* Pulsante quando lo seleziono */
.chatType-buttons button.active {
  background-color: #007bff;
  color: #fff;
  border-color: #007bff;
}

.input-area {
  display: flex;
  flex-direction: column;
  padding: 10px;
  border-top: 1px solid #ddd;
  background-color: #f9f9f9;
  width: 100%;
}

.input-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.message-input {
  flex: 1;
  padding: 8px;
  border-radius: 8px;
  border: 1px solid #ccc;
}

.send-button {
  background-color: #007bff;
  color: white;
  padding: 8px 14px;
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.file-label {
  padding: 20px;
  cursor: pointer;
  font-size: 15px;
}
.file-label-message {
  padding: 20px;
  cursor: pointer;
  font-size: 15px;
}

.image-preview {
  margin-bottom: 100px;
  display: flex;
  align-items: center;
}

.image-preview img {
  width: 200px; 
  height: auto; 
  max-width: 150px; 
  max-height: 150px;
}

.image-preview-upload {
  margin-bottom: 40px;
  display: flex;
  align-items: center;
}

.image-preview-upload img {
  width: 200px; 
  height: auto; 
  max-width: 150px; 
  max-height: 150px;
}

.remove-image-btn {
  background: transparent;
  border: none;
  font-size: 16px;
  cursor: pointer;
}

.message-content img {
  max-width: 200px;
  max-height: 200px;
  object-fit: cover;
  border-radius: 4px;
}



</style>
