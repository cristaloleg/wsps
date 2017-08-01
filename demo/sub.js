$(function() {
    if (!window["WebSocket"]) {
        return;
    }

    var content = $("#content");
    var connSub = new WebSocket('ws://localhost:3001/ws');

    connSub.onopen = function(e) {
        content.attr("disabled", true);

        var m = {
            topic: "default",
        };
        console.log(JSON.stringify(m));
        connSub.send(JSON.stringify(m));
    };

    connSub.onclose = function(e) { 
        content.attr("disabled", true);
    };

    connSub.onmessage = function(e) {
        var m = JSON.parse(e.data);
        console.log(m);
        content.val(content.val() + '\n' + m.body);
    };
});