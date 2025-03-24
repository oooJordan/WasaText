<template>
  <div class="login-wrapper">
    <h1 class="title">WASATEXT</h1>
    <div class="login-container">
      <h2>Login</h2>
      <form @submit.prevent="doLogin">
        <div class="form-group">
          <label for="username">Username:</label>
          <input type="text" id="username" v-model="username" required />
        </div>
        <button type="submit">Login</button>
      </form>
      <p v-if="errormsg" class="error">{{ errormsg }}</p>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      username: '',
      errormsg: null,
    };
  },
  methods: {
    async doLogin() {
      try {
        const response = await fetch(`${__API_URL__}/session`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ name: this.username }),
        });
        if (!response.ok) {
          throw new Error("Login failed");
        }
        
        const data = await response.json();
        if(!data.user_id || data.user_id <= 0){
          throw new Error("Invalid user ID");
        }

        localStorage.setItem('token', String(data.user_id));
        localStorage.setItem('username', this.username);
        this.$router.push('/chat');
      } catch (error) {
        console.error("Error logging in:", error);
        this.errormsg = error.message;
      }
        
    }
  }
};
</script>

<style scoped>
.login-wrapper {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f0f2f5;
}

.title {
  margin-bottom: 1em;
  color: #007bff;
  font-size: 3.5em;
}

.login-container {
  max-width: 400px;
  width: 100%;
  padding: 2em;
  border: 1px solid #ccc;
  border-radius: 8px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  background-color: #fff;
  text-align: center;
}

h2 {
  margin-bottom: 1em;
  color: #333;
}

.form-group {
  margin-bottom: 1.5em;
}

label {
  display: block;
  margin-bottom: 0.5em;
  color: #555;
}

input[type="text"] {
  width: 100%;
  padding: 0.5em;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-sizing: border-box;
}

button {
  padding: 0.75em 1.5em;
  border: none;
  border-radius: 4px;
  background-color: #007bff;
  color: white;
  font-size: 1em;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}

.error {
  margin-top: 1em;
  color: red;
}
</style>