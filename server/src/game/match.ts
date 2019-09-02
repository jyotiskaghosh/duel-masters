import events from "./events"
import { IPlayer, setupPlayer } from "./player"
import shortid from "shortid"
import Phase from "./phase"
import WebSocket from "ws"
import { sendError, sendChooseDeck, sendWarning, sendStateUpdate } from "../net/responses"
import User, { IUser } from "../models/user"
import Deck from "../models/deck"
import { getClientAttachments } from "../net/server"
import { ICard } from "./cards/types"
import { getCard } from './cards/repository'
import { stateUpdateFx, setPhaseFx, resolveSummoningSicknessFx, untapFx } from "./effects"

export interface IMatch {
    id: string,
    inviteId: string,
    name: string,
    description: string,
    phase: Phase,
    player1?: IPlayer,
    player2?: IPlayer,
    playerTurn?: IPlayer
}

let matches: Array<IMatch> = []

export const createMatch = (host: string, name: string, description: string): IMatch => {

    let match = {
        id: shortid.generate(),
        inviteId: shortid.generate(), 
        name,
        description,
        phase: Phase.IDLE
    }

    matches.push(match)

    return match

}

export const getMatch = (id: string) => {
    return matches.find(x => x.id === id)
}

export const addPlayer = async (client: WebSocket, user: IUser, matchId: string, inviteId: string) => {

    let match = getMatch(matchId)

    if(!match) {
        return sendError(client, "Match is no longer available")
    }

    if(match.phase !== Phase.IDLE) {
        return sendError(client, "Match is currently in progress")
    }

    if(inviteId !== match.inviteId) {
        return sendError(client, "Invite id does not match")
    }

    if(match.player1 && match.player2) {
        return sendError(client, "Both players have already connected")
    }

    let decks = await Deck.find().or([{ owner: user.uid }, { standard: true }]).select('uid name cards standard public -_id')

    let player: IPlayer = {
        user,
        client,
        match,
        decks,
        hand: [],
        shieldzone: [],
        manazone: [],
        graveyard: [],
        battlezone: []
    }

    getClientAttachments(client).player = player

    if(!match.player1) {

        match.player1 = player

    } else {

        match.player2 = player

        match.phase = Phase.CHOOSE_DECK
        sendChooseDeck(match.player1.client, match.player1.decks)
        sendChooseDeck(match.player2.client, match.player2.decks)

    }

}

export const playerChooseDeck = async (player: IPlayer, deckId: string) => {

    if(player.match.phase !== Phase.CHOOSE_DECK) {
        return
    }

    if(player.deck) {
        return
    }

    let deck = player.decks.find(x => x.uid === deckId)

    if(!deck) {
        return sendWarning(player.client, "You do not have the rights to use that deck")
    }

    player.deck = createDeck(deck.cards)

    for(let card of player.deck) {
        card.virtualId = shortid()
        card.tapped = false
        card.setup(player.match, player)
    }

    tryStartMatch(player.match)

}

export const createDeck = (cardsIds: string[]): ICard[] => {

    let deck: ICard[] = []

    for(let cardId of cardsIds) {

        let cardInstance: ICard = { ...getCard(cardId) }
        deck.push(cardInstance)

    }

    return deck

}

export const tryStartMatch = (match: IMatch): boolean => {

    if(!match.player1 || !match.player1.deck) {
        return false
    }

    if(!match.player2 || !match.player2.deck) {
        return false
    }

    setupPlayer(match.player1)
    setupPlayer(match.player2)

    match.playerTurn = (Math.random() > 0.5) ? match.player1 : match.player2

    stateUpdateFx([match.player1, match.player2])

    beginTurn(match)

    return true

}

// Step 1: Begin your turn
// Resolve any summoning sickness from creatures in the battle zone.
const beginTurn = (match: IMatch) => {

    setPhaseFx(match, Phase.BEGIN_TURN_STEP)

    for(let creature of match.playerTurn.battlezone) {
        resolveSummoningSicknessFx(creature)
    }

    untapStep(match)

}

// Step 2: Untap step
// Your creatures in the battle zone and cards in your mana zone are untapped. This is forced.
const untapStep = (match: IMatch) => {

    setPhaseFx(match, Phase.UNTAP_STEP)

    for(let creature of match.playerTurn.battlezone) {
        untapFx(creature)
    }

    for(let card of match.playerTurn.manazone) {
        untapFx(card)
    }

    stateUpdateFx([match.player1, match.player2])

    startTurnStep(match)

}

// Step 3: Start of turn step
// Any abilities that trigger at "the start of your turn" are resolved now.
const startTurnStep = (match: IMatch) => {

    setPhaseFx(match, Phase.START_TURN_STEP)

    // TODO: dispatch turn-start event

}

export const before = <K extends keyof events>(eventName: K, listener: (event: events[K], next: Function) => void) => {

}

export const after = <K extends keyof events>(eventName: K, listener: (event: events[K], next: Function) => void) => {

}