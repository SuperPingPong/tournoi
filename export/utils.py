START_LINE = 7


def get_cells_to_update(worksheet):
    # Get the range of cells to update
    player_ids = worksheet.range(f'A{START_LINE}:A')
    license_numbers = worksheet.range(f'B{START_LINE}:B')
    surname = worksheet.range(f'D{START_LINE}:D')
    name = worksheet.range(f'E{START_LINE}:E')
    club = worksheet.range(f'F{START_LINE}:F')
    rank = worksheet.range(f'G{START_LINE}:G')
    category = worksheet.range(f'H{START_LINE}:H')
    tournament_tables_day_1 = worksheet.range(f'K{START_LINE}:Q')
    tournament_tables_day_2 = worksheet.range(f'S{START_LINE}:Y')
    emails = worksheet.range(f'AC{START_LINE}:AC')
    return (
        player_ids,
        license_numbers,
        surname,
        name,
        club,
        rank,
        category,
        tournament_tables_day_1,
        tournament_tables_day_2,
        emails
    )


def clean_worksheet(worksheet):
    (
        player_ids,
        license_numbers,
        surname,
        name,
        club,
        rank,
        category,
        tournament_tables_day_1,
        tournament_tables_day_2,
        emails
    ) = get_cells_to_update(worksheet)
    cells_to_update = player_ids + license_numbers + surname + name + club + rank + category \
        + tournament_tables_day_1 + tournament_tables_day_2 + emails

    # Update the cells with an empty string
    for cell in cells_to_update:
        cell.value = ''

    # Update the cells in bulk
    worksheet.update_cells(cells_to_update)


def fill_worksheet(worksheet, bands, entries):
    # set remaining
    bands = {
        name: {
            'day': band['day'],
            'index': band['index'],
            'remaining': band['max_entries'],
        } for name, band in bands.items()
    }

    choice_mapping = {}
    for key, item in enumerate(entries):
        entry = dict(item)
        permit_id = entry['permit_id']
        band_name = entry['band_name']

        if choice_mapping.get(permit_id) is None:
            choice_mapping[permit_id] = {
                'email': entry.get('email'),
                'last_name': entry.get('last_name'),
                'first_name': entry.get('first_name'),
                'club_name': entry.get('club_name'),
                'points': entry.get('points'),
                'category': entry.get('category'),
                'bands': {},
            }

        bands[band_name]['remaining'] -= 1
        remaining_after = bands[band_name]['remaining']
        if remaining_after >= 0:
            choice_mapping[permit_id]['bands'][band_name] = 1
        else:
            choice_mapping[permit_id]['bands'][band_name] = f'L{abs(remaining_after)}'

    (
        player_ids,
        license_numbers,
        surname,
        name,
        club,
        rank,
        category,
        tournament_tables_day_1,
        tournament_tables_day_2,
        emails
    ) = get_cells_to_update(worksheet)
    cells_to_update = player_ids + license_numbers + surname + name + club + rank + category \
        + tournament_tables_day_1 + tournament_tables_day_2 + emails

    length_bands_day_1 = len([name for name, band in bands.items() if band['day'] == 1])
    length_bands_day_2 = len([name for name, band in bands.items() if band['day'] == 2])

    for key, (permit_id, choice_values) in enumerate(choice_mapping.items()):
        player_ids[key].value = 1 + key
        license_numbers[key].value = permit_id
        emails[key].value = choice_values.get('email')
        surname[key].value = choice_values.get('last_name')
        name[key].value = choice_values.get('first_name')
        club[key].value = choice_values.get('club_name')
        rank[key].value = choice_values.get('points')
        category[key].value = choice_values.get('category')
        license_numbers[key].value = permit_id

        for band_name, cell_value in choice_mapping[permit_id]['bands'].items():
            if bands[band_name]['day'] == 1:
                prefix = length_bands_day_1
                band_index = bands[band_name]['index']
                tournament_table = tournament_tables_day_1
            else:
                prefix = length_bands_day_2
                band_index = bands[band_name]['index'] - length_bands_day_1
                tournament_table = tournament_tables_day_2
            tournament_table[prefix * key + band_index].value = cell_value

    # Update the cells in bulk
    worksheet.update_cells(cells_to_update)
