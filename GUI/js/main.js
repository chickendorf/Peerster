$("#sendBtn").click(function(){
	//var title=$(this).attr("value");
	//alert("SEND");

	$.post("/sendMessage/", {newMessage : $("#messageContent").val()}, function(){
		$("#messageContent").val("")
	})
});


$("#addBtn").click(function(){
	$.post("/addPeer/", {newPeer : $("#newPeer").val()}, function(){
		$("#newPeer").val("")
	})
});

setInterval(function(){
	$.getJSON("/getDatas/", function(data){
		//console.log(data)
		$("#peerList").empty();
		$("#listMessage").empty();
		$("#IDBox").text(data.gID[0]);

		$.each(data.peers, function(index, element) {
		  $("#peerList").append('<li>' + element + '</li>');
		});

		$.each(data.messages, function(index, element) {
		  $("#listMessage").append('<li>' + element + '</li>');
		});

	})
}, 1000)
