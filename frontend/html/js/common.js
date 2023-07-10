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
      checkboxE.checked = true;
    }
  }

  const checkboxes = $('input[type="checkbox"]');

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

  checkboxes.each(function () {
    const checkbox = $(this);

    if (checkboxTarget.getAttribute('id') === checkbox.attr('id')) {
      return
    }
    if (checkbox.attr('data-day') !== checkboxTarget.getAttribute('data-day')) {
      return
    }
    if (checkbox.attr('data-color') !== checkboxTarget.getAttribute('data-color')) {
      return
    }

    const label = $('label[for="' + checkbox.attr('id') + '"]');

    if (checkboxTarget.checked === true) {
      checkbox.prop('disabled', true);
      label.attr('data-title', 'Vous ne pouvez pas sélectionner deux tableaux de la même couleur');
    }
    if (checkboxTarget.checked === false) {
      checkbox.prop('disabled', false);
      label.removeAttr('data-title');
    }
  })
}
