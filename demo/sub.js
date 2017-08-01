$(function() {
    if (!window["WebSocket"]) {
        return;
    }

    var content = $("#content");
    var connSub = new WebSocket('ws://localhost:3001/ws');

    connSub.onopen = function(e) {
        content.attr("disabled", false);
    };

    connSub.onclose = function(e) { 
        content.attr("disabled", true);
    };

    connSub.onmessage = function(e) {
        if (e.data != content.val()) {
            content.val(e.data);
        }
    };
});