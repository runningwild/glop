**This file may become out-of-date with respect to the files it describes.  Any time this file references other files you should look at those files if you want to be sure to have the most up-to-date information**


Actions represent anything that an Entity can do in the game.  Available actions are listed in this directory, one per file, with the prefix "action_".  With the exception of action_move_utils.go this gives the following Actions:

    aoe
    basic attack
    chain attack
    charge attack
    counter attack
    move
    multistrike
    linear shift
    spray attack


Each action is specified by zero or more parameters that define how it works in the game.  For example the AoE action is defined with the following parameters:

  * Cost - Ap required to use this ability
  * Range - How far away the center of the effect can be from the Entity using this Action
  * Size - The radius of the AoE
  * Allies - Non-zero if the AoE can affect allies
  * Enemies - Non-zero if the AoE can affect enemies
  * Effects - List of effects applied to anyone in the AoE

Actions are loaded from json files.  Here is a sample stanza that could be added to a json file that gets loaded with RegisterAllSpecsInDir().

    {
      "Type" : "aoe",
      "Name" : "Grenade",
      "Icon_path" : "weapons/grenade.png",
      "Effects" : [
        "Shrapnel"
      ],
      "Int_params" : {
        "Cost" : 1,
        "Range" : 6,
        "Size" : 2,
        "Allies" : 1,
        "Enemies" : 1
      }
    }

This creates an Action called 'Grenade'.  The Grenade Action costs 1 Ap to use, can target a cell up to a distance of 6 away from the Entity using it, has a radius of 2, 
and affects all Entities in the AoE with the Shrapnel effect, which presumably does damage.

As an example of what else can be done with the same kind of Action, here is another example of an AoE:

    {
      "Type" : "aoe",
      "Name" : "Defender's Aura",
      "Icon_path" : "weapons/defender.png",
      "Effects" : [
        "Major Shield",
        "Minor Heal"
      ],
      "Int_params" : {
        "Cost" : 4,
        "Range" : 0,
        "Size" : 2,
        "Allies" : 1,
        "Enemies" : 0
      }
    }

This describes an Action called "Defender's Aura", which is an AoE that must be centered on the Entity using it, has a radius of 2, only affects allies, and gives them two effects: Major Shield and Minor Heal.

As of right now, here are all Actions and their associated parameters:

# "aoe"
* **Cost** - Ap required to use this action
* **Range** - Maximum distance from the acting Entity to the target
* **Size** - Radius of the AoE
* **Allies** - 1 if this affects allies, 0 otherwise
* **Enemies** - 1 if this affects enemies, 0 otherwise
* **Effects** - List of names of effects to apply to anything affect by the AoE

# "basic attack"
* **Cost** - Ap required to use this action
* **Range** - Maximum distance from the acting Entity to the target
* **Power** - Attack power
* **Melee** - 1 if this action should use a melee attack animation, 0 for a ranged attack animation

# "chain attack"
* **Cost** - Ap required to use this action
* **Range** - Maximum distance from the acting Entity to the target
* **Power** - Attack power
* **Melee** - 1 if this action should use a melee attack animation, 0 for a ranged attack animation
* **Adds** - Number of *additional* attacks this action can generate

# "charge attack"
(There is no cost for a charge attack - the cost is whatever the movement cost would be for that same path)
* **Power** - Attack power

# "counter attack"
* **Cost** - Ap required to use this action
* **Range** - Distance an enemy unit needs to be for this interrupt to trigger
* **Power** - Attack power
* **Melee** - 1 if this action should use a melee attack animation, 0 for a ranged attack animation

# "move"
(There is no cost for a move - the cost is whatever the movement cost would be for that path)

# "multistrike"
* **Cost** - Ap required to use this action
* **Range** - Maximum distance from the acting Entity to the target
* **Power** - Attack power
* **Melee** - 1 if this action should use a melee attack animation, 0 for a ranged attack animation
* **Count** - Maximum number of targets

# "linear shift"
* **Cost** - Ap required to use this action
* **Range** - Maximum distance from the acting Entity to the target
* **Pull** - Maximum distance the target can be pulled
* **Push** - Maximum distance the target can be pushed

# "spray attack"
* **Cost** - Ap required to use this action
* **Power** - Attack power
* **Melee** - 1 if this action should use a melee attack animation, 0 for a ranged attack animation
* **Length** - Length of the spray attack
* **Start** - Width of the spray attack on the side closest to the acting unit is Start*2 - 1
* **End** - Width of the spray attack on the side furthest from the acting unit is End*2 - 1
