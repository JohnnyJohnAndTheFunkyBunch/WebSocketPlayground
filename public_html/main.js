    var colors = ["red", "aqua", "green", "yellow", "fuchsia", "gray", "lime", "maroon", "navy", "olive"];
    var playerBox = $('.player-box');
    var consoleBox = $('.console-box');
    var contentBox = $('.content');
    var host = "ws://localhost:8000";
    var idMsg = "{\"T\":0,\"M\":\""+session+"\"}";
    var playerId;
    console.log(idMsg);
    console.log("Host:", host);
    
    var s = new WebSocket(host);
    
    s.onopen = function (e) {
        console.log("Socket opened.");
        $('.alert').hide();
        $('.alert-warning').show();
        s.send(idMsg);
    };
    
    s.onclose = function (e) {
        console.log("Socket closed.");
        $('.alert').hide();
        $('.alert-danger').show();
        s.send("{\"T\":1}");
    };
    
    s.onmessage = function (e) {
        console.log(e.data);
        var obj = JSON.parse(e.data);
        var type = obj.T;
        if (type == 0) {
            addPlayer(obj.P, "Player" + (obj.P).toString());
            addMessage('Player' + obj.P + ' connected to session: ', colors[obj.P % 11]);
        }
        else if (type == 1) {
            var name = $('#' + obj.P).find('.name').text();
            addMessage(name + ' disconnected from session', colors[obj.P % 11]);
            removePlayer(obj.P);
        }
        else if (type == 2) {
            $('.alert').hide();
            $('.alert-success').show();
            playerId = obj.P;
            addOwnPlayer(playerId);
            addMessage('Connected to session: ' + session, "black");
            addOtherPlayers(obj.S.Players);
            changeContent(obj.S.C)
        }
        else if (type == 4) {
            $('#' + obj.P).find('.latency').text((obj.L/1000000).toFixed(1) + 'ms');
        }
        else if (type == 5) {
            var name = $('#'+obj.P).find('.name').text();
            addMessage(name + ' changed the content', colors[obj.P % 11]);
            changeContent(obj.C);
        }
        else if (type == 6) {
            onApplicationMsg(obj);
        }
    };
    
    s.onerror = function (e) {
        console.log("Socket error:", e);
    };

    $('#mainmenu').click(function () {
        var msg = '{"P":'+playerId+',"T":5,"C":"main"}';
        s.send(msg);
    });

    function addPlayer(id, username) {
        htmlArray = [];
        htmlArray.push('<div id="'+id+'" class="player">');
        htmlArray.push('<div class="square" style="background:' + colors[id % 11] + ';"><span class="playerid">'+id+'</span></div>');
        htmlArray.push('<div class="name">'+username+'</div>');
        htmlArray.push('<div class="latency"></div>');
        htmlArray.push('</div>');
        playerBox.append(htmlArray.join(''));
    }
    function addOwnPlayer(id) {
        htmlArray = [];
        htmlArray.push('<div id="'+id+'" class="player">');
        htmlArray.push('<div class="square" style="background:' + colors[id % 11] + ';"><span class="playerid">ME</span></div>');
        htmlArray.push('<div class="name">Player'+id+'</div>');
        htmlArray.push('<div class="latency"></div>');
        htmlArray.push('</div>');
        playerBox.append(htmlArray.join(''));
    }
    function addMessage(msg, color) {
        consoleBox.append('<span style="color:' + color + ';">' + msg +'</span><br>');
        consoleBox[0].scrollTop = consoleBox[0].scrollHeight;
    }
    function removePlayer(id) {
        $('#'+id).remove();
    }
    function addOtherPlayers(players) {
        if (players == null) {
            return;
        }
        var size = players.length;
        for (var i = 0 ; i < size ; i++) {
            addPlayer(players[i].Id, players[i].Name);
        }
    }
    function changeContent(content) {
        contentBox.html(content);
    }



