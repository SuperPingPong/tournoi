function Survey(survey) {
  if (!survey) {
    throw new Error("No Form Survey found!");
  }

  // select the elements
  const progressbar = survey.querySelector(".progressbar");
  const surveyPanels = survey.querySelectorAll(".survey__panel");
  const question1Email = survey.querySelector("[name='email']");
  const question2License = survey.querySelector("[name='license']");
  const question3CheckBoxes = survey.querySelectorAll("[name='question_3']");
  const allPanels = Array.from(survey.querySelectorAll(".survey__panel"));
  let progressbarStep = Array.from(progressbar.querySelectorAll(".progressbar__step "));
  const mainElement = document.querySelector("main");
  const nextButton = survey.querySelector("[name='next']");
  const prevButton = survey.querySelector("[name='prev']");
  const submitButton = survey.querySelector("[name='submit']");
  let currentPanel = Array.from(surveyPanels).filter(panel => panel.classList.contains("survey__panel--current"))[0];
  const formData = {};
  const options = {
    question1Email,
    question2License,
    question3CheckBoxes,
  };
  let dontSubmit = false;

  function storeInitialData() {
    submitButton.disabled = true;
    submitButton.setAttribute("aria-hidden", true);
    allPanels.map(panel => {
      let index = panel.dataset.index;
      let panelName = panel.dataset.panel;
      let question = panel.querySelector(".survey__panel__question").textContent.trim();
      formData[index] = {
        panelName: panelName,
        question: question
      };
    });
  }

  function updateProgressbar() {
    let index = currentPanel.dataset.index;
    let currentQuestion = formData[`${parseFloat(index)}`].question;
    progressbar.setAttribute("aria-valuenow", index);
    progressbar.setAttribute("aria-valuetext", currentQuestion);
    progressbarStep[index - 1].classList.add("active");
  }

  function updateFormData({ target }) {
    const index = +currentPanel.dataset.index;
    const { name, type, value } = target;
    checkRequirements();

    formData[index].answer = {
      [name]: value
    };
  }

  function showError(input, text) {
    const formControl = input.parentElement;
    const errorElement = formControl.querySelector(".error-message");
    errorElement.innerText = text;
    errorElement.setAttribute("role", "alert");
    if (survey.classList.contains("form-error")) return;
    survey.classList.add("form-error");
  }

  function noErrors(input) {
    if (!input) {
      const errorElement = currentPanel.querySelector(".error-message");
      errorElement.textContent = "";
      errorElement.removeAttribute("role");
      survey.classList.remove("form-error");
      return;
    }
    const formControl = input.parentElement;
    const errorElement = formControl.querySelector(".error-message");
    errorElement.innerText = "";
    errorElement.removeAttribute("role");
  }

  function checkEmail(input) {
    if (input.value.trim() === "") {
      showError(input, `Le champ email est obligatoire`);
    } else {
      const pattern = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
      if (pattern.test(input.value.trim())) {
        noErrors(input);
      } else {
        showError(input, "Le format de l'email n'est pas valide.");
      }
    }
  }

  function checkLicense(input) {
    if (input.value.trim() === "") {
      showError(input, `Le champ licence est obligatoire`);
    } else {
      const suggestions = $('#suggestions');
      if (!suggestions.is(':visible')) {
        const numberRegex = /^[A-Z0-9]+$/;
        if (numberRegex.test(input.value.trim())) {
          noErrors(input);
        } else {
          showError(input, "Le format du numéro de licence n'est pas valide.");
        }
      }
    }
  }

  function checkMemberBands(checkboxes) {
    let isAtLeastOneChecked = false;
    for (let i = 0; i < checkboxes.length; i++) {
      if (checkboxes[i].checked) {
        isAtLeastOneChecked = true;
        break;
      }
    }
    if (isAtLeastOneChecked) {
      noErrors(survey.querySelector('.survey__panel__hearabout'));
    } else {
      showError(
        survey.querySelector('.survey__panel__hearabout'),
        "Au moins un tableau doit être sélectionné"
      );
    }
 }

  function checkRequirements() {
    const requirement = currentPanel.dataset.requirement;
    const index = currentPanel.dataset.index;
    const errorElement = currentPanel.querySelector(".error-message");

    if (index === "1") {
      checkEmail(question1Email);
    }
    if (index === "2") {
      checkLicense(question2License);
    }
    if (index === "3") {
      const question3CheckBoxes = survey.querySelectorAll("[name='question_3']");
      checkMemberBands(question3CheckBoxes);
    }
    if (survey.classList.contains("form-error")) {
      // errorElement.textContent = `Le champ ${requirement} est invalide.`;
      errorElement.setAttribute("role", "alert");
      survey.classList.add("form-error");
    }

  }

  function updateProgressbarBar() {
    const index = currentPanel.dataset.index;
    let currentQuestion = formData[`${parseFloat(index)}`].question;
    progressbar.setAttribute("aria-valuenow", index);
    progressbar.setAttribute("aria-valuetext", currentQuestion);
    progressbarStep[index].classList.remove("active");
  }

  function displayNextPanel() {
    currentPanel.classList.remove("survey__panel--current");
    currentPanel.setAttribute("aria-hidden", true);
    currentPanel = currentPanel.nextElementSibling;
    currentPanel.classList.add("survey__panel--current");
    currentPanel.setAttribute("aria-hidden", false);
    updateProgressbar();
    if (+currentPanel.dataset.index > 2) {
      prevButton.disabled = false;
      prevButton.setAttribute("aria-hidden", false);
    }
    if (+currentPanel.dataset.index === 2) {
      if (isAfterDeadline()) {
        $('.no-switch').hide();
        $('.switch').show();
        prevButton.disabled = true;
        prevButton.setAttribute("aria-hidden", true);
        nextButton.disabled = true;
        nextButton.setAttribute("aria-hidden", true);
        submitButton.disabled = true;
        submitButton.setAttribute("aria-hidden", true);
      } else {
        $('.no-switch').show();
        $('.switch').hide();
      }
    }
    if (+currentPanel.dataset.index === 3) {
      nextButton.disabled = true;
      nextButton.setAttribute("aria-hidden", true);
      submitButton.disabled = false;
      submitButton.setAttribute("aria-hidden", false);
    }
    if (+currentPanel.dataset.index === 3) {
      // Dynamically field form for set-entries endpoint
      const memberId = localStorage.getItem('memberId');
      const memberPoints = localStorage.getItem('memberPoints');
      const memberSex = localStorage.getItem('memberSex');
      $.ajax({
        url: `/api/members/${memberId}/band-availabilities`,
        type: 'GET',
        contentType: 'application/json',
        success: function(response) {
          const sessionId = response.session_id;
          [1, 2].forEach(day => {
            let bandsDay = response.bands.filter(band => band.Day === day);
            let bandDayContainer = document.getElementById(`form-group-day-${day}`);
            bandDayContainer.innerHTML = '';
            bandsDay.forEach(band => {
              const div = document.createElement('div');
              div.classList.add('form-group');
              const input = document.createElement('input');
              input.type = 'checkbox';
              input.classList.add('checkbox');
              input.id = `tableau-${band.Name}`;
              input.name = 'question_3';
              input.value = band.ID;
              input.setAttribute('data-color', band.Color);
              input.setAttribute('data-day', band.Day);
              input.setAttribute('data-maxpoints', band.MaxPoints);
              input.setAttribute('data-sex', band.Sex);
              input.setAttribute('data-member-points', memberPoints);
              input.setAttribute('data-member-sex', memberSex);
              const label = document.createElement('label');
              label.htmlFor = `tableau-${band.Name}`;
              label.textContent = `Tableau ${band.Name} (${band.MaxPoints >= 9000 ? 'TC' : '≤ ' + band.MaxPoints + ' pts'}) - ` +
                `${band.Available > 0 ? band.Available + " place(s) restante(s)" : ""}` +
                `${band.Available === 0 ? "Inscription en liste d'attente" : ""}`;
              div.appendChild(input);
              div.appendChild(label);
              bandDayContainer.appendChild(div);
            });
            // Create hidden input for sessionId
            const hiddenInput = document.createElement('input');
            hiddenInput.type = 'hidden';
            hiddenInput.name = 'sessionId';
            hiddenInput.value = sessionId;
            bandDayContainer.appendChild(hiddenInput);
          });
          const question3CheckBoxes = survey.querySelectorAll("[name='question_3']");
          question3CheckBoxes.forEach(elem => elem.addEventListener("click", manageCheckboxRequisitesEvent));
        },
        error: function(xhr, textStatus, error) {
          notificationError();
        }
      });
    }
  }

  function displayPrevPanel() {
    currentPanel.classList.remove("survey__panel--current");
    currentPanel.setAttribute("aria-hidden", true);
    currentPanel = currentPanel.previousElementSibling;
    currentPanel.classList.add("survey__panel--current");
    currentPanel.setAttribute("aria-hidden", false);
    updateProgressbarBar();
    if (+currentPanel.dataset.index <= 2) {
      prevButton.disabled = true;
      prevButton.setAttribute("aria-hidden", true);
    }
    if (+currentPanel.dataset.index < 3) {
      nextButton.disabled = false;
      nextButton.setAttribute("aria-hidden", false);
      submitButton.disabled = true;
      submitButton.setAttribute("aria-hidden", true);
    }
  }

  function handleprevButton() {
    displayPrevPanel();
  }

  function ask_otp(email) {
    // Make API query to send OTP code via email
    $.ajax({
      url: '/api/otp',
      type: 'POST',
      data: JSON.stringify({ email: email }),
      contentType: 'application/json',
      success: function(response) {
        // Show SweetAlert2 popup to ask for OTP code
        Swal.fire({
          title: 'Entrer le dernier code OTP reçu',
          html: '<span>Code OTP envoyé sur: ' + email + '<br><span style="font-size: 75%">(<i>N\'oubliez pas de vérifier vos spams</i><span>)</span>',
          input: 'text',
          showCancelButton: true,
          confirmButtonText: 'Confirmer',
          cancelButtonText: 'Annuler',
          confirmButtonColor: '#5468D4',
          cancelButtonColor: '#dc3741',
          showLoaderOnConfirm: true,
          preConfirm: function(code) {
            // Verify OTP code
            return $.ajax({
              url: '/api/login',
              type: 'POST',
              data: JSON.stringify({ email: email, secret: code }),
              contentType: 'application/json',
              timeout: 3000, // Add timeout option to abort the request after 3 seconds
              error: function(xhr, textStatus, error) {
                // Handle error if OTP code is invalid
                if (xhr.status == 401) {
                  notificationError('', 'Le code OTP est incorrect');
                } else {
                  notificationError();
                }
              }
            });
          },
          allowOutsideClick: false
        }).then(function(result) {
          // Handle success after OTP code is verified
          if (result.isConfirmed) {
            Swal.fire({
              icon: 'success',
              title: 'Code validé',
              text: 'Le code OTP a été vérifié avec succès',
              // confirmButtonText: 'OK'
              showConfirmButton: false,
              timer: 1500
            }).then(function(result) {
              // Go to next panel
              if (result.isConfirmed || true) {
                noErrors();
                displayNextPanel();
              }
            });
          }
        });
      },
      error: function(xhr, textStatus, error) {
        // Handle error if API query fails
        if (xhr.status == 400) {
          Swal.fire('Adresse email invalide');
        } else {
          notificationError();
        }
      }
    });
  }

  function handleNextButton(event) {
    survey.classList.remove("form-error");
    const index = currentPanel.dataset.index;
    // console.log(formData[index]);

    // ignore step 1 and step 2
    /*
    if (index === "1" || index === "2") {
      noErrors();
      displayNextPanel();
      return;
    }
    */

    checkRequirements();
    if (survey.classList.contains("form-error")) {
      return;
    }

    if (index === "1") {
      var email = $('input[name="email"]').val().toLowerCase();

      // Test if jwt is valid
      $.ajax({
        url: '/api/check-auth',
        type: 'POST',
        data: JSON.stringify({ email: email }),
        contentType: 'application/json',
        success: function(response) {
          if (response.valid === true) {
            Swal.fire({
              icon: 'success',
              title: "Validation de votre adresse email",
              text: "L'adresse email est validée",
              confirmButtonText: 'OK',
              showConfirmButton: false,
              timer: 1500
            }).then(function(result) {
              // Go to next panel
              if (result.isConfirmed || true) {
                noErrors();
                displayNextPanel();
              }
            });
          } else {
            ask_otp(email.toLowerCase().trim());
          }
        },
        error: function(xhr, textStatus, error) {
          // console.log(error);
          ask_otp(email.toLowerCase().trim());
        }
      });
    }

    if (index === "2") {
      var licenseNumber = $('input[name="license"]').val();
      // Handle the Enter key press here, e.g., run your search function.
      if (event.isTrusted === false) {
        /*
        isTrusted set to true is typically generated by a user action
        isTrusted set to false is likely programmatically created.
        */
        event.preventDefault();
        return
      }
      $.ajax({
        url: '/api/players/' + licenseNumber,
        type: 'GET',
        data: {
          license_number: licenseNumber
        },
        success: function(response) {
          // console.log(response);
          // Generate HTML content based on the AJAX response
          const htmlContent = '<p>' + response.content + '</p>';
          // Show SweetAlert2 with HTML content
          Swal.fire({
            title: 'Confirmer les informations',
            html:
                '👤 Nom: ' + response.nom + '<br>' +
                '👤 Prénom: ' + response.prenom + '<br>' +
                '🧾 N° License: ' + response.licence + '<br>' +
                '🗂 Catégorie: ' + response.cat + '<br>' +
                '🏓 Club: ' + response.nomclub.replace(' ', ' ') + '<br>' +
                '⚧ Sexe: ' + response.sexe + '<br>' +
                '🎯 Officiels: ' + response.point ,
            showCancelButton: true,
            confirmButtonText: 'Confirmer',
            cancelButtonText: 'Annuler',
            confirmButtonColor: '#5468D4',
            cancelButtonColor: '#dc3741',
            allowOutsideClick: false
          }).then(function(result) {
            if (!result.isConfirmed) {
              return
            }
            // Create member if not exists or pass if exists but no set-entries done yet
            $.ajax({
              url: '/api/members',
              type: 'GET',
              data: {
                permit_id: licenseNumber
              },
              contentType: 'application/json',
              success: function(response) {
                let members = response.Members;
                if (members.length === 0) {
                  $.ajax({
                    url: '/api/members',
                    type: 'POST',
                    contentType: 'application/json',
                    data: JSON.stringify({ permitid: licenseNumber }),
                    success: function(response) {
                      let memberId = response.ID
                      localStorage.setItem('memberId', memberId);
                      localStorage.setItem('memberPoints', response.Points);
                      localStorage.setItem('memberSex', response.Sex);
                      noErrors();
                      displayNextPanel();
                    },
                    error: function(xhr, textStatus, error) {
                      if (xhr.status === 409) {
                        Swal.fire({
                          icon: 'error',
                          title: 'Licencié déjà inscrit au tournoi',
                          html: 'Le joueur est déjà inscrit par un autre utilisateur. Contactez nous.',
                        });
                      } else {
                        notificationError();
                      }
                    }
                  });
                } else {
                  let selectedMember = members[0];
                  let memberBands = selectedMember.Entries;
                  if (memberBands === null || memberBands.length === 0) {
                    let memberId = selectedMember.ID
                    localStorage.setItem('memberId', memberId);
                    noErrors();
                    displayNextPanel();
                  } else {
                    Swal.fire({
                      icon: 'error',
                      title: 'Licencié déjà inscrit au tournoi',
                      html: 'Le joueur est déjà inscrit. Vous pouvez modifier votre inscription ou consulter votre rang dans les listes d\'attentes: ' +
                      '<a href="/app">Cliquez ici</a>',
                    });
                  }
                }
              },
              error: function(xhr, textStatus, error) {
                notificationError();
              }
            });
         });
         },
         error: function(xhr, textStatus, error) {
           notificationError();
         }
      });
    }

    return

    if (!formData[index].hasOwnProperty("answer")) {
      checkRequirements();
    } else {
      noErrors();
      displayNextPanel();
    }
  }

  // submitting the form
  function handleFormSubmit(e) {
    e.preventDefault();
    survey.classList.remove("form-error");
    const index = currentPanel.dataset.index;
    //  console.log(formData[index]);
    checkRequirements();
    if (survey.classList.contains("form-error")) {
      return;
    }
    // const index = currentPanel.dataset.index;
    const memberId = localStorage.getItem('memberId');
    const bandIDs = [];
    $('input[type="checkbox"]:checked').each(function() {
      bandIDs.push($(this).val());
    });
    const sessionId = $('input[name="sessionId"]').val();
    $.ajax({
      url: '/api/bands',
      type: 'GET',
      contentType: 'application/json',
      success: function(response) {
        let checkboxStringTitles = [
          '<p>Samedi 28 Octobre 2023</p>', '<p>Dimanche 29 Octobre 2023</p>'
        ]
        const bands = response.bands.filter(band => bandIDs.includes(band.ID));
        let confirmText = '';
        [1, 2].forEach(day => {
          let bandsDayCreatedItems = [];
          let bandsDayCreated = bands.filter(band => band.Day === day);
          bandsDayCreated.forEach(band => {
            bandsDayCreatedItems.push(`<p style="text-align: left; margin: 0">✅ Ajout du tableau ${band.Name}</p>`)
          })
          if (bandsDayCreatedItems.length > 0) {
            confirmText += checkboxStringTitles[day-1] + bandsDayCreatedItems.join('')
          }
        })
        $.ajax({
          url: `/api/members/${memberId}`,
          type: 'GET',
          contentType: 'application/json',
          success: function(member) {
            // console.log(member);
            Swal.fire({
              title: 'Confirmer les tableaux',
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
                return bands
              }
            }).then((result) => {
              if (result.isConfirmed) {
                $.ajax({
                  url: `/api/members/${memberId}/set-entries`,
                  type: 'POST',
                  contentType: 'application/json',
                  data: JSON.stringify({
                    bandids: bandIDs,
                    sessionid: sessionId,
                  }),
                  success: function(response) {
                    // console.log(response);
                    if (result.isConfirmed) {
                      mainElement.classList.add("submission");
                      mainElement.setAttribute("role", "alert");
                      mainElement.innerHTML = `<div style="margin: 2.5em auto; text-align: center"><svg width="126" height="118" fill="none" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 126 118" aria-hidden="true"><path d="M52.5 118c28.995 0 52.5-23.729 52.5-53S81.495 12 52.5 12 0 35.729 0 65s23.505 53 52.5 53z" fill="#B9CCED"/><path d="M45.726 87L23 56.877l8.186-6.105 15.647 20.74L118.766 0 126 7.192 45.726 87z" fill="#A7E9AF"/></svg></div>
                      <h2 class="submission">Merci pour votre inscription</h2>
                      <p style="text-align: center">Une confirmation vous a été envoyé par email.
                      <br><span style="font-size: 88%">(<i>N'oubliez pas de vérifier vos spams</i><span>)</span>
                      <p style="text-align: center">
                          <a href="/" style="margin-top: 0.5em; background-color: #5f6caf; color: #ffffff; display: inline-block; text-decoration: none; font-family: Helvetica, Arial, sans-serif; font-size: 14px; line-height: 1.4; text-align: center; padding: 10px 25px; border-radius: 3px;">
                              Revenir au menu principal
                          </a>
                          <a href="/app" style="margin-top: 0.5em; background-color: #5f6caf; color: #ffffff; display: inline-block; text-decoration: none; font-family: Helvetica, Arial, sans-serif; font-size: 14px; line-height: 1.4; text-align: center; padding: 10px 25px; border-radius: 3px;">
                              Gérer mes inscriptions
                          </a>
                      </p>
                      `;
                      return false;
                    }
                  },
                  error: function(xhr, textStatus, error) {
                    notificationError();
                  }
                });
              }
            });
          },
          error: function(xhr, textStatus, error) {
            notificationError();
          }
        });
      },
      error: function(xhr, textStatus, error) {
        notificationError();
      }
    });
  }

  storeInitialData();

  // Add event listeners
  function addListenersTo({ question1Email, question2License, question3CheckBoxes}) {
    question1Email.addEventListener("change", updateFormData);
    question2License.addEventListener("change", updateFormData);
    question3CheckBoxes.forEach(elem => elem.addEventListener("change", updateFormData));
  }
  nextButton.addEventListener("click", handleNextButton);
  prevButton.addEventListener("click", handleprevButton);
  addListenersTo(options);
  survey.addEventListener("submit", handleFormSubmit);

  function manageIfAuth() {
    const index = currentPanel.dataset.index;
    $.ajax({
      url: '/api/members',
      type: 'GET',
      contentType: 'application/json',
      success: function(response) {
        noErrors();
        displayNextPanel();
      },
      error: function(xhr, textStatus, error) {
        // console.log(error);
      }
    });
  }

  manageIfAuth();

}

const survey = Survey(document.querySelector(".survey"));
