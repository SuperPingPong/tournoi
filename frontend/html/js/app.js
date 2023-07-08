function init() {
  $.ajax({
    url: '/api/members',
    type: 'GET',
    success: function(response) {
      let filteredMembers = response.Members.filter(member => member.Entries !== null);
      if (response.Members.length === 0) {
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
      let dataTable = document.getElementById("dataTable");
      dataTable.style.display = "block";
      dataTable = $('#dataTable').DataTable({
        "lengthMenu": [10, 25, 50, 100],
        "pageLength": 10,
        "data": filteredMembers,
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
              const bandNames = 'Tableaux ' + row.Entries.map(entry => entry.BandName).join(' / ');
              return bandNames;
            }
          },
          {
            data: null,
            render: function(data, type, row) {
              const editButton = '<div class="field"><button type="submit" data-info=\'' + JSON.stringify(row) + '\'><i class="fa-solid fa-pencil"></i></button></div>';
              return editButton;
            }
          }
        ],
        "initComplete": function() {
          // Attach click event listener to buttons
          $('button[type="submit"]').on('click', function(event) {
            event.preventDefault();
            const member = $(this).attr('data-info');
            editMemberBands(member);
          });
        }
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

function editMemberBands(memberString) {
  const member = JSON.parse(memberString);
  var checkboxStrings = ['', '']
  let checkboxStringTitles = [
    'Samedi 28 Octobre 2023', 'Dimanche 29 Octobre 2023'
  ]
  $.ajax({
    url: '/api/bands',
    type: 'GET',
    contentType: 'application/json',
    success: function(response) {
      [1, 2].forEach(day => {
        let bandsDay = response.bands.filter(band => band.Day === day);
        bandsDay.forEach(band => {
          checkboxStrings[day-1] += `<div class="form-group" style="text-align: left"><input type="checkbox" class="checkbox" id="tableau-${band.Name}" name="editMemberBands" value="${band.ID}"><label for="tableau-${band.Name}">Tableau ${band.Name} (72 places restantes)</label></div>`;
        })
      })
      Swal.fire({
          title: 'Mise a jour des tableaux',
          html:
            '👤 Nom: ' + member.LastName + ' | ' +
            '👤 Prénom: ' + member.FirstName + ' | ' +
            '🧾 N° License: ' + member.PermitID + ' | ' +
            '🗂️ Catégorie: ' + member.Category + ' | ' +
            '🏓 Club: ' + member.ClubName.replace(' ', ' ') + ' | ' +
            '⚧ Sexe: ' + member.Sex + ' | ' +
            '🎯Officiels: ' + member.Points + '<br><br>' +
            checkboxStringTitles[0] + '<br><br>' + checkboxStrings[0] + '<br>' +
            checkboxStringTitles[1] + '<br><br>' + checkboxStrings[1] + '<br>',
          // input: 'text',
          inputAttributes: {
              autocapitalize: 'off'
          },
          showCancelButton: true,
          showLoaderOnConfirm: true,
          confirmButtonText: 'Mettre a jour',
          cancelButtonText: 'Annuler',
          confirmButtonColor: '#5468D4',
          cancelButtonColor: '#dc3741',
          preConfirm: (data) => {
            // to complete
          },
          allowOutsideClick: () => !Swal.isLoading()
      }).then((result) => {
          if (result.isConfirmed) {
            // to complete
          }
      })
    },
    error: function(xhr, textStatus, error) {
      Swal.fire({
        icon: 'error',
        title: 'Une erreur est survenue',
        text: ''
      });
    }
  });
}

$(document).ready(function() {
  init();
});
