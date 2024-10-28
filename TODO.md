# ToDo

### Changement de tournoi
- Pensez à rechercher les mots clefs suivant pour remplacement  des informations des tableaux (grep -Ri):
- - 2024 + 2024-12 + décembre
- - 'const targetDateTime'
- - 'Les inscriptions seront accessibles ici'
- - féminin
- - band-day png + reglement pdf
- - 'tableaux maximum par jour'
- - '2 tableaux'
- - 'checkedCheckboxesWithSameDay.length >= 2'
- - 'checkboxE' (tableau féminin) / tableau-E / 'tableau E'
- - TODO
- - 'Aucun tableau disponible pour les joueurs supérieurs à 1999 points'
Pensez à changer le secret jwt sur une nouvelle version de l'appli avant lancement pour forcer l'invalidation des precedents jwt
Avant le lancement du tournoi le redirect vers /announcement doit être géré au niveau nginx et pas dans le front sinon probleme de non refresh lié au cache des navigateurs au lancement
```
location = / {
    try_files /farewell.html =404;
}
```
Si jamais un window.location a été mis dans un script, on doit forcer la revalidation du coté client:
```
location = / {
    add_header Cache-Control "no-cache, no-store, must-revalidate";
    try_files /index.html =404;
}
```

Pour forcer le rechargement des scripts js, ajouter un get param, exemple:
```
<script src="js/common.js?v=202410"></script>
```

A tester également avec nginx:
```
location ~* \.(js)$ {
    add_header Cache-Control "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0";
    add_header Pragma "no-cache";
    expires off;
}
```

Une autre idée c'est de changer le nom de domaine à chaque fois avec la date dans le nom de domaine

### Features:
- Gérer la modification de date de création de certaines entries pour gérer les inscriptions faites par mail si bug/probleme pour acceder à l'application en prenant en compte la date de réception du mail

- Rajouter une ligne d'explication dans le front pour dire que les rangs dans la liste d'attente peut evoluer dans les deux sens si desinscription ou inscription faite a posteriori pour ceux ayant envoyé un mail et ayant un probleme technique en tenant compte de la date de leur demande par email

- Rajouter un msg après l'arrêt des inscriptions dans la partie /app pour signifier qu'il n'est plus possible de modifier
- ~~Rajouter une page html avec un message de remerciement et un lien vers le site de lognestt quand le tournoi ets terminé~~
- ~~Disable event enter (sentry errors) input on search player~~
- Notifier les members lorsqu'un player n'est plus en liste d'attente
--> container séparé: WIP
- Afficher le rang dans la liste d'attente au moment de l'inscription (>En fait quand un tableau est plein et que tu veux tu inscrire, ce serait bien que tu saches directement combien tu seras et ne pas le découvrir au dernier moment. Voire même ce serait bien avant de mettre ton adresse mail, de savoir quels tableaux sont remplis comment)
- ~~Reparer la logique de disable checkbox sur la partie update~~
- Authentification par OTP à changer, avec qqch de full front user/password or otp (magic)

### Informations à communiquer FAQ:
- Communiquer sur la partie buvette / menus avec prix
- Pour la FAQ : Que si c'est bon pour la liste d'attente, on ne sait pas, que sur certains tableaux, on a pris des gens qui étaient N°30 sur la liste d'attente et d'autres ou en n'a pas pris un seul
- Rappeler le système des tickets pour la buvette, qu'on rembourse s'il leur reste des tickets donc ne pas hésiter à prévoir plus d'argent ça ne sera jamais perdu
- Préciser que le paiement se fait au jour le jour même si on fait samedi et dimanche, cheque ou espèces
- Bien dire pas de CB, que le DAB le plus proche est à la gare RER de Lognes
- De prévenir si jamais on ne vient pas
- Qu'à l'heure pile de fin de pointage (heure du gymnase qui fait foi), on scratche les joueurs qui ne se sont pas présentés et qui n'ont pas prévenu
- En rappelant l'adresse de contact \<censored\>
- Il est toujours utile de prévenir des désistements même après la fin des inscriptions, même le jour même

~~### To fix:~~
~~- bug player wong nathalie detected as M in api endpoint search player~~
~~--> pas réussi a reproduire le bug~~
- ~~Penser à faire la formule pour le nombre de présents (c'est écrit "72", pas (=somme...)~~
- ~~fix mail providers in error (orange.fr, laposte.net, wanadoo.fr)~~
