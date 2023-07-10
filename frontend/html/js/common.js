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