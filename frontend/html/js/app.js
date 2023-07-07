function init() {
  $.ajax({
    url: '/api/members',
    type: 'GET',
    success: function(response) {
      if (response.members.length === 0) {
        Swal.fire({
          icon: 'error',
          title: "Modification des tableaux",
          text: "Aucune inscription n'est associé à votre compte, vous allez être redirigé vers la page d'accueil pour enregistrer votre première inscription.",
          confirmButtonText: 'OK',
          showConfirmButton: true,
          timer: 5000,
        }).then(function() {
          window.location.href = '/';
        });
      }
      console.log(response.members);
      response.members.forEach(member => {
        console.log(member);
      });
      let dataTable = document.getElementById("dataTable");
      dataTable.style.display = "block";
      dataTable = $('#dataTable').DataTable({
        "lengthMenu": [10, 25, 50, 100],
        "pageLength": 10,
        "data": response.members,
        "order": [], // Remove default sorting
        "columns": [
          {
            data: null,
            render: function(data, type, row) {
              return `👤 ${row.FirstName} ${row.LastName}`;
            }
          },
          {
            data: null,
            render: function(data, type, row) {
              return `🏓 ${row.ClubName}`;
            }
          },
          {
            data: null,
            render: function(data, type, row) {
              return `🎯 ${row.Points}`;
            }
          },
          {
            data: null,
            render: function(data, type, row) {
              const bandNames = 'Tableaux ' + row.Bands.map(band => band.Name).join(' / ');
              return bandNames;
            }
          },
          { "defaultContent": "<div class=\"field\"><button type=\"submit\"><i class=\"fa-solid fa-pencil\"></i></button></div>" },
        ],
      });
    },
    error: function(xhr, textStatus, error) {
      // console.log(error);
      Swal.fire({
        icon: 'error',
        title: "Echec de l'authentification",
        text: "Vous n'êtes plus authentifié ou votre authentification a expiré, vous allez être redirigé vers la page de connexion",
        confirmButtonText: 'OK',
        showConfirmButton: true,
        timer: 5000,
      }).then(function() {
        window.location.href = '/';
      });
    }
  });
}

$(document).ready(function() {
  init();
});
