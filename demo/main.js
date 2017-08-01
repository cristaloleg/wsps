$(function() {
    if (!window["WebSocket"]) {
        return;
    }

    var content = $("#content");
    var connPub = new WebSocket('ws://localhost:3000/ws');
    var connSub = new WebSocket('ws://localhost:3002/ws');

    connPub.onopen = function(e) {
        content.attr("disabled", false);
    };

    connPub.onclose = function(e) { 
        content.attr("disabled", true);
    };

    connSub.onmessage = function(e) {
        if (e.data != content.val()) {
            content.val(e.data);
        }
    };

    var timeoutId = null;
    // var typingTimeoutId = null;
    // var isTyping = false;

    // content.on("keydown", function() {
    //     isTyping = true;
    //     window.clearTimeout(typingTimeoutId);
    // });

    content.on("keyup", function() {
        // typingTimeoutId = window.setTimeout(function() {
        //     isTyping = false;
        // }, 1000);

        window.clearTimeout(timeoutId);
        timeoutId = window.setTimeout(function() {
            // if (isTyping) return;

            var text = content.val().replace(/\n/g,'');

            var m = {
                topic:"default",
                body:text,
            };
    
            console.log(m);
            console.log(JSON.stringify(m));

            connPub.send(JSON.stringify(m));

        }, 1100);
    });
});