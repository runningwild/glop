**This file may become out-of-date with respect to the files it describes.  Any time this file references other files you should look at those files if you want to be sure to have the most up-to-date information**

An Entity is defined by a set of stats that can refer to attributes and actions defined elsewhere.  The stats that can be defined in an Entity definition are:

* Name - A name that can be used to refer to this Entity
* Health
* Ap
* Defense
* LosDist - Maximum distance that this unit can see, reduced by certain kinds of terrain
* Atts - List of attribute names
* Sprite - Name of the sprite used for this Entity
* Weapons - List of action names

Here is a sample Entity definition, most values should be self-explanitory.

    {
      "Name" : "Cheetah",
      "Health" : 3,
      "Ap" : 6,
      "Defense" : 2,
      "LosDist" : 15,
      "Atts" : [
        "standard",
        "fleet"
      ],
      "Sprite" : "green",
      "Weapons" : [
        "Cheetah Bite",
        "Pounce"
      ]
    }

