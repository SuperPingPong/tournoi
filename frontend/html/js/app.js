let dataTable;

function initDataTable() {
  let dataTableHTML = document.getElementById("dataTable");
  dataTableHTML.style.display = "block";
  dataTable = $('#dataTable').DataTable({
    "lengthMenu": [15, 30, 60, 100],
    "pageLength": 15,
    "serverSide": true,
    "ajax": {
      "url": "/api/members",
      "dataSrc": function (data) {
        data.recordsTotal = data.Total;
        data.recordsFiltered = data.Total;
        return data.Members;
      },
      "data": function (params) {
        params.search = $('input[aria-controls="dataTable"]').val();
        params.page = (params.start / params.length) + 1;
        params.page_size = params.length;
        return params
      },
    },
    "order": [],
    "language": {
        "url": '/locales/datatable/fr-FR.json',
    },
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
          const bandNames = row.Entries === null ? '' : row.Entries.map(entry => entry.BandName).join(' / ');
          return bandNames;
        }
      },
      {
        data: null,
        render: function(data, type, row) {
          const editButton = '<button type="submit" data-action="edit" data-info=\'' + JSON.stringify(row) + '\'><i class="fa-solid fa-pencil"></i></button>';
          const deleteButton = '<button type="submit" data-action="delete" data-info=\'' + JSON.stringify(row) + '\'><i class="fa-solid fa-times" style="color: red;"></i></button>';
          const buttonsContainer = '<div class="field">' + editButton + deleteButton + '</div>';
          return buttonsContainer;
        }
      }
    ],
    "initComplete": function() {
      let searchInput = $('input[aria-controls="dataTable"]');
      searchInput.on('keyup', function () {
        dataTable.search(this.value).draw();
      });
      // Attach click event listener to buttons
      $('button[data-action="edit"]').on('click', function(event) {
        event.preventDefault();
        const member = $(this).attr('data-info');
        editMemberBands(member);
      });
      $('button[data-action="delete"]').on('click', function(event) {
        event.preventDefault();
        const member = $(this).attr('data-info');
        deleteMember(member);
      });
    }
  });
}

function init() {
  initDataTable();
  return;
  $.ajax({
    url: '/api/members',
    type: 'GET',
    success: function(response) {
      const members = response.Members;
      if (members.length === 0) {
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
      initDataTable();
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
  const bandIDs = member.Entries === null ? [] : member.Entries.map(obj => obj.BandID);
  var checkboxStrings = ['', '']
  let checkboxStringTitles = [
    '<p>Samedi 28 Octobre 2023</p>', '<p>Dimanche 29 Octobre 2023</p>'
  ]
  $.ajax({
    url: `/api/members/${member.ID}/band-availabilities`,
    type: 'GET',
    contentType: 'application/json',
    success: function(response) {
      const sessionId = response.session_id;
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
            // console.log(result);
            if (result.isConfirmed) {
              $.ajax({
                url: `/api/members/${member.ID}/set-entries`,
                type: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({
                  bandids: result.value,
                  sessionid: sessionId,
                }),
                success: function(response) {
                  Swal.fire({
                    icon: 'success',
                    title: 'Mise à jour effectuée',
                    text: '',
                    showConfirmButton: false,
                    timer: 3000,
                  }).then((result) => {
                    // force reload dataTable
                    $.ajax({
                      url: '/api/members',
                      type: 'GET',
                      success: function(response) {
                        const members = response.Members;
                        dataTable.destroy();
                        initDataTable();
                        // dataTable.clear().rows.add(members).draw();
                      },
                      error: function(xhr, textStatus, error) {
                        Swal.fire({
                          icon: 'error',
                          title: 'Une erreur est survenue',
                          text: ''
                        });
                      }
                    });
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

function deleteMember(memberString) {
  const member = JSON.parse(memberString);
  console.log(member);
  Swal.fire({
    title: "Suppression l'inscription",
    html:
      '👤 Nom: ' + member.LastName + ' | ' +
      '👤 Prénom: ' + member.FirstName + ' | ' +
      '🧾 N° License: ' + member.PermitID + ' | ' +
      '🗂️ Catégorie: ' + member.Category + ' | ' +
      '🏓 Club: ' + member.ClubName.replace(' ', ' ') + ' | ' +
      '⚧ Sexe: ' + member.Sex + ' | ' +
      '🎯Officiels: ' + member.Points + '<br><br>' +
      "Êtes-vous certain de vouloir supprimer l'inscription de ce joueur ?",
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
  }).then((result) => {
    if (result.isConfirmed) {
      console.log(member);
      Swal.fire({
        icon: 'success',
        title: 'Suppression effectuée',
        text: ''
      }).then(() => {
        console.log(member);
      });
    }
  });
}

$(document).ready(function() {
  init();
});
