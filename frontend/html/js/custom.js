function search() {
    const input = $('#license');
    const value = input.val();
    // var surname = value.split(" ")[0];
    // var name = value.split(" ")[1] || "";
    var surname = value;
    var name = "";

    // Regular expression to match only numbers
    const numberRegex = /^[0-9]+$/;
    if (numberRegex.test(value)) {
      // Do not make query on /api/players if license number is copy/pasted
      return
    }

    $.ajax({
      url: "/api/players",
      type: "POST",
      data: JSON.stringify({
        surname: surname,
        name: name
      }),
      contentType: 'application/json',
      success: function(players) {
        const suggestions = $('#suggestions');
        const result = $('#result');
        suggestions.html("");
        for (const player of players.players) {
          const div = $('<div>').html(player.nom + ' ' + player.prenom + ' - ' + player.nomclub + ' - ' + player.point);
          div.click(function() {
            input.val("");
            suggestions.hide();
            input.val(player.licence);
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
