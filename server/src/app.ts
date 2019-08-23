import { createServer } from "http"
import logger from "./utils/logger"
import api from "./api"
import * as server from "./net/server"
import mongoose from "mongoose"

export const bootstrap = async () => {

    logger.info("Starting up...")

    // db
    mongoose.connect(process.env.MONGO_URI, { useNewUrlParser: true })

    // http
    const web = createServer(api)

    // ws
    server.connect(web)

    web.listen(process.env.PORT || 3000, () => {
        logger.info(`Listening on port ${process.env.PORT || 3000}`)
    })

} 