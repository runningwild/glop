**This file may become out-of-date with respect to the files it describes.  Any time this file references other files you should look at those files if you want to be sure to have the most up-to-date information**


Attributes are a way of defining a set of basic stats that can be applied to any unit.  For example you can define an attribute called 'Arboreal' for a unit that gets a bonus to attack when standing in forest, then give the Arboreal attribute to any unit you want.  This attribute can be changed later and the change will apply to all units with the Arboreal attribute.

Attributes can define any or all of the following:

* LosMods
* MoveMods
* DefenseMods
* AttackMods

For each of the categories listed above you can list one or more Terrain/Mod pairs that indicate the value that that should be used for that stat when dealing with that particular terrain, specifically:

* LosMods - The LoS cost required to see past a cell of a particular terrain
* MoveMods - The Ap cost required to move through a cell of a particular terrain
* AttackMods - The attack value modifier when in a cell of a particular terrain
* DefenseMods - The defense value modifier when in a cell of a particular terrain

If there are conflicting attributes they are resolved as follows:

* LosMods - The lowest value is used
* MoveMods - The lowest value is used
* AttackMods - The last value is used
* DefenseMods - The last value is used

If a particular terrain is not specified in LosMods then it blocks line of site.  If a particular terrain is not specified in MoveMods then it is impassable.

Consider the following example:

    {
      "standard" : {
        "LosMods" : {
          "jungle" : 2,
          "plains" : 0,
          "water" : 0,
        },
        "MoveMods" : {
          "jungle" : 1,
          "plains" : 0,
        },
        "DefenseMods" : {
          "jungle" : 1,
        },
        "AttackMods" : {
          "hills" : 1
        }
      },
      "jungle dweller" : {
        "LosMods" : {
          "jungle" : 1
        },
        "MoveMods" : {
          "jungle" : 0
        },
        "AttackMods" : {
          "jungle" : 1
        }
      },
      "fleet" : {
        "AttackMods" : {
          "plains" : 1
        }
      },
      "aquatic" : {
        "MoveMods" : {
          "water" : 0
        },
        "LosMods" : {
          "water" : 0
        },
      }
    }

In this case 'standard' indicates a typical unit that can move over land but not water.  It comes with some simple modifiers for attack and defense as well.  'aquatic' indicates a unit that can move and see through water *only*.  A unit that is ['standard', 'aquatic'] can move through plains, jungle, and water.  A unit that is ['standard', 'jungle dweller'] can move and see through jungle terrain more easily than a 'standard' unit, and also gets an attack bonus while in jungle terrain.
