var gamemode = document.getElementById("pvp")
var botOptions = document.getElementById("bot")

gamemode.addEventListener('change', function() {
	if (this.value) {
		botOptions.style.display = 'none'
	} else {
		botOptions.style.display = 'block'
	}
})
