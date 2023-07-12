let dataTable;

function initDataTable() {
  let dataTableHTML = document.getElementById("dataTable");
  dataTableHTML.style.display = "block";
  dataTable = $('#dataTable').DataTable({
    "lengthMenu": [5, 10, 25, 50],
    "pageLength": 5,
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
          const bandPrices = row.Entries === null ? '' : row.Entries.map(entry => entry.BandPrice);
          const sumOfBandPrices = bandPrices.length > 0 ? bandPrices.reduce((total, price) => total + price) : '';
          return sumOfBandPrices;
        }
      },
      {
        data: null,
        render: function(data, type, row) {
          const historyButton = '<button style="display:none" type="submit" data-action="history" data-info=\'' + JSON.stringify(row) + '\'><i class="fa-solid fa-history"></i></button>';
          const mailButton = '<button style="display: none" type="submit" data-action="mail" data-info=\'' + JSON.stringify(row) + '\'><i class="fa-solid fa-envelope"></i></button>';
          const editButton = '<button type="submit" data-action="edit" data-info=\'' + JSON.stringify(row) + '\'><i class="fa-solid fa-pencil"></i></button>';
          const deleteButton = '<button type="submit" data-action="delete" data-info=\'' + JSON.stringify(row) + '\'><i class="fa-solid fa-rectangle-xmark" style="color: red;"></i></button>';
          const buttonsContainer = '<div class="field">' + historyButton + mailButton + editButton + deleteButton + '</div>';
          return buttonsContainer;
        }
      }
    ],
    "drawCallback": function(settings) {
      // Attach click event listener to parent element (dataTable)
      const isAdmin = settings.json.IsAdmin
      if (isAdmin === true) {
        $('button[data-action="history"]').show();
        $('#dataTable').off('click', 'button[data-action="history"]').on('click', 'button[data-action="history"]', function(event) {
          event.preventDefault();
          const memberString = $(this).attr('data-info');
          historyMemberBands(memberString);
        });
        $('button[data-action="mail"]').show();
        $('#dataTable').off('mail', 'button[data-action="mail"]').on('click', 'button[data-action="mail"]', function(event) {
          event.preventDefault();
          const memberString = $(this).attr('data-info');
          mailMemberBands(memberString);
        });
      }
    },
    "initComplete": function() {
      let searchInput = $('input[aria-controls="dataTable"]');
      searchInput.on('keyup', function () {
        dataTable.search(this.value).draw();
      });
      // Attach click event listener to parent element (dataTable)
      $('#dataTable').on('click', 'button[data-action="edit"]', function(event) {
        event.preventDefault();
        const memberString = $(this).attr('data-info');
        editMemberBands(memberString);
      });
      $('#dataTable').on('click', 'button[data-action="delete"]', function(event) {
        event.preventDefault();
        const memberString = $(this).attr('data-info');
        deleteMember(memberString);
      });
    }
  });
}

function formatEventDetails(event) {
  let emoji = '';

  if (event.EventByIsAdmin) {
    emoji = '🛡️';
  }

  switch (event.EventType.toLowerCase()) {
    case 'created':
      emoji += '✅ Ajout du tableau ' + event.BandName;
      break;
    case 'deleted':
      emoji += '❌ Suppression du tableau ' + event.BandName;
      break;
  }

  const eventTime = new Date(event.EventTime);
  const eventTimeFormatted = eventTime.toLocaleString('fr-FR', {
    timeZone: 'Europe/Paris',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });

  return `${emoji} - ${eventTimeFormatted}`;
}

function historyMemberBands(memberString) {
  const member = JSON.parse(memberString);
  console.log(member);
  $.ajax({
    url: `/api/members/${member.ID}/get-entries-history`,
    type: 'GET',
    contentType: 'application/json',
    success: function(response) {
      const history = response.history;
      let historyText = ''
      if (history) {
        history.forEach(event => {
          historyText += formatEventDetails(event) + '<br>';
        })
      }
      Swal.fire({
        title: 'Historique des modifications',
        html:
          '👤 Nom: ' + member.LastName + ' | ' +
          '👤 Prénom: ' + member.FirstName + ' | ' +
          '🧾 N° License: ' + member.PermitID + ' | ' +
          '🗂️ Catégorie: ' + member.Category + ' | ' +
          '🏓 Club: ' + member.ClubName.replace(' ', ' ') + ' | ' +
          '⚧ Sexe: ' + member.Sex + ' | ' +
          '🎯Officiels: ' + member.Points + '<br><br>' +
          '<div style=""><h4>📄 Légende 📄</h4><div>🛡️: Évènement effectué par un admin</div></div><br><br>' +
          '<div style="text-align: left">' + historyText + '</div>',
        showCancelButton: false,
        confirmButtonText: 'OK',
        confirmButtonColor: '#5468D4'
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

function mailMemberBands(memberString) {
  const member = JSON.parse(memberString);
  Swal.fire({
    title: 'Email du joueur',
    html: `<a href="mailto:${member.User.UserEmail}">${member.User.UserEmail}</a>`,
    icon: 'info',
    showCancelButton: false,
    confirmButtonText: 'OK',
    confirmButtonColor: '#5468D4',
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
          checkboxStrings[day-1] += `<div class="form-group" style="text-align: left">` +
            `<input type="checkbox" ${bandIDs.includes(band.ID) ? "checked" : ""} ` +
            `class="checkbox" id="tableau-${band.Name}" ` +
            `data-color="${band.Color}" data-day="${band.Day}"` +
            `data-member-points="${member.Points}" data-member-sex="${member.Sex}"` +
            `data-maxpoints="${band.MaxPoints}" data-sex="${band.Sex}"` +
            `data-member="${member.ID}" name="editMemberBands" value="${band.ID}">` +
            `<label for="tableau-${band.Name}">` +
             `Tableau ${band.Name} (${band.MaxPoints >= 9000 ? 'TC' : '≤ ' + band.MaxPoints + ' pts'}) - ` +
                `${band.Available >= 0 ? band.Available + " place(s) restante(s)" : ""}` +
                `${band.Available === 0 ? "Inscription en liste d'attente" : ""}` +
            `</label>` +
            `</div>`;
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
           '<div class="rules-container"><h2>⚠️ Règlement ⚠️</h2><ul><li>Les tableaux de couleurs identiques ne pourront pas être cumulés dans la même journée.</li><li>Les féminines ont une participation dans le tableau « E » (Féminin ≤ 1199pts) obligatoire (pour le samedi uniquement, si les conditions sont remplies).</li><li>3 tableaux maximum par jour.</li><li>Les inscriptions pourront se faire jusqu’au vendredi 27 octobre 2023 – 12H00.</li><li>Les places disponibles sont bloquées pendant 10 minutes, au-delà votre session sera expirée.</li></ul></div><br><br>' +
          checkboxStringTitles[0] + checkboxStrings[0] +
          checkboxStringTitles[1] + checkboxStrings[1],
        // input: 'text',
        customClass: 'custom-swal-html-container',
        didRender: () => {
          const checkboxes = document.querySelectorAll('input[type="checkbox"]');
          checkboxes.forEach(checkbox => {
            checkbox.addEventListener('click', manageCheckboxRequisitesEvent);
            if (checkbox.checked) {
              manageCheckboxRequisites(checkbox);
            }
          });
        },
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
      $.ajax({
        url: `/api/members/${member.ID}`,
        type: 'DELETE',
        success: function(response) {
          Swal.fire({
            icon: 'success',
            title: 'Suppression effectuée',
            text: '',
            showConfirmButton: false,
            timer: 3000,
          }).then((result) => {
            dataTable.destroy();
            initDataTable();
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
  });
}

function init() {
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
          return
        });
      }
      const isAdmin = response.IsAdmin;
      if (isAdmin === true) {
        $('p[id="export"]').show();
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

  // Add click event listener to logout button
  $('#logoutButton').on('click', function(event) {
    event.preventDefault();
    Swal.fire({
      title: 'Confirmation de déconnexion',
      text: 'Êtes-vous sûr de vouloir vous déconnecter ?',
      icon: 'question',
      showLoaderOnConfirm: true,
      showCancelButton: true,
      confirmButtonText: 'Confirmer',
      cancelButtonText: 'Annuler',
      confirmButtonColor: '#5468D4',
      cancelButtonColor: '#dc3741',
    }).then((result) => {
      if (result.isConfirmed) {
        window.location.href = '/';
      }
    });
  });

  $('#exportButton').on('click', function (event) {
    event.preventDefault();
    Swal.fire({
      title: 'Exporter les données',
      text: 'Êtes-vous sûr de vouloir télécharger les données ?',
      icon: 'warning',
      showLoaderOnConfirm: true,
      showCancelButton: true,
      confirmButtonText: 'Confirmer',
      cancelButtonText: 'Annuler',
      // confirmButtonColor: '#5468D4',
      confirmButtonColor: '#1F7145',
      cancelButtonColor: '#dc3741',
    }).then((result) => {
      if (result.isConfirmed) {
        // window.location.href = '/api/export';
        Swal.fire({
          title: 'Traitement en cours',
          html: 'Veuillez patienter...<br><progress value="0" max="10"></progress>',
          timer: 10 * 1000, // timer in ms
          timerProgressBar: true,
          allowOutsideClick: false,
          didOpen: () => {
            const progressBar = Swal.getHtmlContainer().querySelector('progress');
            const interval = setInterval(() => {
              const currentValue = parseInt(progressBar.value);
              if (currentValue < 10) {
                progressBar.value = currentValue + 1;
              }
            }, 1000);
            Swal.showLoading();
            // Make AJAX call to /api/export
            // Replace the below line with your own logic
            fetch('/api/export')
              .then(response => {
                clearInterval(interval);
                if (response.ok) {
                  return response.blob();
                } else {
                  throw new Error('Une erreur est survenue lors du téléchargement.');
                }
              })
              .then(blob => {
                const downloadLink = document.createElement('a');
                const url = URL.createObjectURL(blob);
                downloadLink.href = url;
                downloadLink.download = 'tournoi-de-lognes.xlsx';
                downloadLink.style.display = 'none';
                document.body.appendChild(downloadLink);

                // Trigger the download
                downloadLink.click();

                // Cleanup the temporary download link
                URL.revokeObjectURL(url);
                document.body.removeChild(downloadLink);

                Swal.fire({
                  title: 'Téléchargement terminé',
                  icon: 'success',
                  // timer: 2000,  // 2 seconds
                  showConfirmButton: true,
                  confirmButtonText: 'OK',
                  cancelButtonText: 'Annuler',
                  // confirmButtonColor: '#5468D4',
                  confirmButtonColor: '#1F7145',
                });
              })
              .catch(error => {
                Swal.fire({
                  title: 'Erreur',
                  text: error.message,
                  icon: 'error',
                  timer: 2000,  // 2 seconds
                  showConfirmButton: false
                });
              });
          }
        });
      }
    });
  });
}


$(document).ready(function() {
  init();
});
