$(function() {
    if (!window["WebSocket"]) {
        return;
    }

    var content = $("#content");
    var connPub = new WebSocket('ws://localhost:3000/ws');

    connPub.onopen = function(e) {
        content.attr("disabled", false);
    };

    connPub.onclose = function(e) { 
        content.attr("disabled", true);
    };

    connPub.onmessage = function(e) {
        if (e.data != content.val()) {
            content.val(e.data);
        }
    };

    var timeoutId = null;

    content.on("keyup", function() {
        window.clearTimeout(timeoutId);
        timeoutId = window.setTimeout(function() {
            var text = content.val().replace(/\n/g,'');
            var d = new Date();
            var now = d.getTime();

            var m = {
                topic: "default",
                body: text,
                time: now,
            };
    
            console.log(m);
            console.log(JSON.stringify(m));

            connPub.send(JSON.stringify(m));

        }, 1100);
    });
});