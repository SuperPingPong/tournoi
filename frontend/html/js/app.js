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
              return `👤 ${row.LastName} ${row.FirstName}`;
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
  const bandIDs = member.Entries.map(obj => obj.BandID);
  var checkboxStrings = ['', '']
  let checkboxStringTitles = [
    '<p>Samedi 28 Octobre 2023</p>', '<p>Dimanche 29 Octobre 2023</p>'
  ]
  $.ajax({
    url: '/api/bands',
    type: 'GET',
    contentType: 'application/json',
    success: function(response) {
      [1, 2].forEach(day => {
        let bandsDay = response.bands.filter(band => band.Day === day);
        bandsDay.forEach(band => {
          checkboxStrings[day-1] += `<div class="form-group" style="text-align: left"><input type="checkbox" ${bandIDs.includes(band.ID) ? "checked" : ""} class="checkbox" id="tableau-${band.Name}" data-member="${member.ID}" name="editMemberBands" value="${band.ID}"><label for="tableau-${band.Name}">Tableau ${band.Name} (72 places restantes)</label></div>`;
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
          checkboxStringTitles[0] + checkboxStrings[0] +
          checkboxStringTitles[1] + checkboxStrings[1],
        // input: 'text',
        inputAttributes: {
            autocapitalize: 'off'
        },
        showCancelButton: true,
        confirmButtonText: 'Mettre a jour',
        cancelButtonText: 'Annuler',
        confirmButtonColor: '#5468D4',
        cancelButtonColor: '#dc3741',
        preConfirm: () => {
          const bands = response.bands.filter(band => bandIDs.includes(band.ID));
          const newBandIDs = $(`[data-member="${member.ID}"]:checked`).map(function() {
            return $(this).val();
          }).get();
          const newBands = response.bands.filter(band => newBandIDs.includes(band.ID));
          // Check if no changes
          if (JSON.stringify(bandIDs.sort()) === JSON.stringify(newBandIDs.sort())) {
            return
          }
          // return diff
          const deletedItems = bands.filter(band => !newBandIDs.includes(band.ID));
          const createdItems = newBands.filter(band => !bandIDs.includes(band.ID));

          let result = {
            'deleted': deletedItems,
            'created': createdItems,
            'newBandIDs': newBandIDs
          }
          return result;
        },
        allowOutsideClick: () => !Swal.isLoading()
      }).then((result) => {
        if (result.isConfirmed && typeof(result.value) !== 'boolean') {
          // Ask for confirmation for the changes
          let confirmText = '';
          [1, 2].forEach(day => {
            let bandsDayUpdated = [];
            let bandsDayCreated = result.value.created.filter(band => band.Day === day);
            bandsDayCreated.forEach(band => {
              bandsDayUpdated.push(`<p style="text-align: left; margin: 0">✅ Ajout du tableau ${band.Name}</p>`)
            })
            let bandsDayDeleted = result.value.deleted.filter(band => band.Day === day);
            bandsDayDeleted.forEach(band => {
              bandsDayUpdated.push(`<p style="text-align: left; margin: 0">❌ Suppression du tableau ${band.Name}</p>`)
            })
            if (bandsDayUpdated.length > 0) {
              confirmText += checkboxStringTitles[day-1] + bandsDayUpdated.join('')
            }
          })
          Swal.fire({
            title: 'Confirmer la mise a jour',
            html:
              '👤 Nom: ' + member.LastName + ' | ' +
              '👤 Prénom: ' + member.FirstName + ' | ' +
              '🧾 N° License: ' + member.PermitID + ' | ' +
              '🗂️ Catégorie: ' + member.Category + ' | ' +
              '🏓 Club: ' + member.ClubName.replace(' ', ' ') + ' | ' +
              '⚧ Sexe: ' + member.Sex + ' | ' +
              '🎯Officiels: ' + member.Points + '<br><br>' +
              confirmText,
            // input: 'text',
            inputAttributes: {
                autocapitalize: 'off'
            },
            showLoaderOnConfirm: true,
            showCancelButton: true,
            confirmButtonText: 'Confirmer',
            cancelButtonText: 'Annuler',
            confirmButtonColor: '#5468D4',
            cancelButtonColor: '#dc3741',
            preConfirm: () => {
              return result.value.newBandIDs
            }
          }).then((result) => {
            console.log(result);
            if (result.isConfirmed) {
              $.ajax({
                url: `/api/members/${member.ID}/set-entries`,
                type: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ bandids: result.value }),
                success: function(response) {
                  Swal.fire({
                    icon: 'success',
                    title: 'Mise à jour effectuée',
                    text: ''
                  });
                  // TODO: force reload dataTable
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
          });
        };
      });
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
