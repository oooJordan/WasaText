<template>
	<router-view/>
</template>

<script>
export default {
	mounted(){
		if(!localStorage.getItem('token')){
			this.$router.push('/login');
		}else{
			this.$router.push('/chat');
		}
	}
}
</script>