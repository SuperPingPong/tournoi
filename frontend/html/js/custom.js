function search() {
    const input = $('#license');
    const value = input.val();
    // var surname = value.split(" ")[0];
    // var name = value.split(" ")[1] || "";
    var surname = value;
    var name = "";
    $.ajax({
      url: "/api/search",
      data: {
        surname: surname,
        name: name
      },
      type: "GET",
      success: function(players) {
        const suggestions = $('#suggestions');
        const result = $('#result');
        suggestions.html("");
        for (const player of players) {
          const div = $('<div>').html(player.surname + ' ' + player.name + ' - ' + player.nclub + ' - ' + player.score);
          div.click(function() {
            input.val("");
            suggestions.hide();
            const license = player.license
            // window.location = '/?license=' + license;
            input.val(player.license);
          });
          suggestions.append(div);
        }
        if (players.length === 0) {
          suggestions.hide();
        } else {
          suggestions.show();
        }
      }
    });
}
