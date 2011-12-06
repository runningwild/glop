**This file may become out-of-date with respect to the files it describes.  Any time this file references other files you should look at those files if you want to be sure to have the most up-to-date information**


Effects represent a specific modification to a unit's stats.  Effects are named and an effect's name is what is used to indicate that an Action applies that effect to another unit.  For example you can create an aoe Action that applies the effect "Fire", then you can define the Fire effect as reducing a unit's Health by 1 every round for 3 rounds.  The rest of this document will describe what can be done with effects, and how to write them.  If a unit with an effect on it has another effect by that same name applied to it, the first effect by that name is overwritten by the second.

Effects are loaded from json files.  Here is a sample stanza that could be added to a json file that gets loaded with RegisterEffectsFromJson().

    {
      "Type" : "static effect",
      "Name" : "Slow",
      "Int_params" : {
        "TimedEffect" : 2,
        "MoveCost" : 1
      }
    }

This creates an Effect called 'Slow'.  This effect lasts for two rounds and causes a unit to need an additional Ap for every cell that it moves through.

As of right now, here are all Effects and their associated parameters:

# "static effect"
* **TimedEffect** - Number of rounds this effect will last
* **Attack** - This value is added to the unit's Attack value
* **Defense** - This value is added to the unit's Defense value
* **LosDist** - This value is added to the unit's maximum LoS distance
* **Health** - This value is added to the unit's health each round (cannot exceed max health)
* **Ap** - This value is added to the unit's Ap each round (*can* exceed max Ap)
* **MoveCost** - This value is added to the move cost of any cell the unit moves through (cannot drop below 1)
* **LosCost** - This value is added to the LoS cost of any cell the unit looks throuw (cannot drop below 1)

# "shield"
* **Amount** - The amount of damage absorbed before this effect dissipates
