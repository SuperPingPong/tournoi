function manageCheckboxRequisitesEvent(event) {
  const checkboxTarget = event.target;
  manageCheckboxRequisites(checkboxTarget)
}

function manageCheckboxRequisites(checkboxTarget) {
  const checkboxE = document.querySelector('input#tableau-E');
  if (checkboxE) {
    if (checkboxE.getAttribute('data-member-sex') === checkboxE.getAttribute('data-sex') &&
      parseInt(checkboxE.getAttribute('data-member-points')) <= parseInt(checkboxE.getAttribute('data-maxpoints'))
    ) {
      console.log(checkboxE);
      const checkboxesWithSameDay = document.querySelectorAll(`input[data-day="${checkboxE.getAttribute('data-day')}"]:checked`);
      console.log(checkboxesWithSameDay);
      if (checkboxesWithSameDay.length > 0) {
        checkboxE.checked = true;
      }
    }
  }

  let checkboxes = $('input[type="checkbox"]');

  const checkboxesWithSameDay = checkboxes.filter(function () {
    const checkbox = $(this);
    return checkbox.attr('data-day') === checkboxTarget.getAttribute('data-day');
  });
  const checkedCheckboxesWithSameDay = checkboxesWithSameDay.filter(':checked');

  if (checkedCheckboxesWithSameDay.length >= 3) {
    checkboxesWithSameDay.each(function () {
      const checkbox = $(this);
      if (!checkbox.is(':checked')) {
        checkbox.prop('disabled', true);
        const label = $('label[for="' + checkbox.attr('id') + '"]');
        label.attr('data-title', 'Vous ne pouvez pas sélectionner plus de 3 tableaux pour cette journée');
      }
    });
  } else {
    checkboxes.each(function () {
      const checkbox = $(this);
      const label = $('label[for="' + checkbox.attr('id') + '"]');
      checkbox.prop('disabled', false);
      label.removeAttr('data-title');
    });
  }

  checkboxes = $('input[type="checkbox"]');
  const checkedCheckboxes = checkboxes.filter(':checked');
  checkedCheckboxes.each(function () {
    const checkedCheckbox = $(this);
    checkedCheckboxColor = checkedCheckbox.attr('data-color')
    checkedCheckboxDay = checkedCheckbox.attr('data-day')
    // Verify if any checkbox in checkedCheckboxes matches the same data-day and data-color
    const conflictingCheckboxes = checkboxes.filter(function () {
      const checkbox = $(this);
      return checkbox.attr('id') !== checkedCheckbox.attr('id') && checkbox.attr('data-day') === checkedCheckboxDay && checkbox.attr('data-color') === checkedCheckboxColor;
    });
    conflictingCheckboxes.each(function () {
      const checkbox = $(this);
      checkbox.prop('disabled', true);
      const label = $('label[for="' + checkbox.attr('id') + '"]');
      label.attr('data-title', 'Vous ne pouvez pas sélectionner deux tableaux de la même couleur');
    });
  });
}


function isAfterDeadline() {
  const parisTime = new Date().toLocaleString('en-US', { timeZone: 'Europe/Paris' });
  const formattedParisTime = new Date(parisTime).toISOString().split('.')[0];
  const targetDateTime = '2023-10-27T12:00:00';
  return formattedParisTime > targetDateTime;
}

function notificationError(text = '', title = 'Une erreur est survenue') {
  Swal.fire({
    icon: 'error',
    title: title,
    text: text
  });
}

function getMemberHeaderHtml(member) {
    return '👤 Nom: ' + member.LastName + ' | ' +
      '👤 Prénom: ' + member.FirstName + ' | ' +
      '🧾 N° License: ' + member.PermitID + ' | ' +
      '🗂️ Catégorie: ' + member.Category + ' | ' +
      '🏓 Club: ' + member.ClubName.replace(' ', ' ') + ' | ' +
      '⚧ Sexe: ' + member.Sex + ' | ' +
      '🎯 Officiels: ' + member.Points + '<br><br>'
}

function logout() {
  $.ajax({
    url: `/api/logout`,
    type: 'GET',
    contentType: 'application/json',
    success: function(response) {
      Swal.fire({
        title: 'Confirmation de déconnexion',
        text: 'Déconnexion effectuée avec succès',
        icon: 'success',
        confirmButtonText: 'OK',
        showConfirmButton: true,
        timer: 5000,
      }).then(function() {
        window.location.href = '/';
        return
      });
    },
    error: function(xhr, textStatus, error) {
      notificationError();
    }
  });
}

function commonInit() {
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
        logout();
      }
    });
  });

  $(document).on("keydown", "#email", function (event) {
    if (event.key === "Enter") {
      event.preventDefault();
      $("[name='next']").trigger("click");
    }
  });
  $(document).on("keydown", "#license", function (event) {
    if (event.key === "Enter") {
      event.preventDefault();
      $("[name='next']").trigger("click");
    }
  });
}

$(document).ready(function() {
  commonInit();
});
