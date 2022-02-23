package irc

// irc commands which may be sent or received by a client.
const (
	CmdAdmin    = "ADMIN"    // Get information about the administrator of a server.
	CmdAway     = "AWAY"     // Set an automatic reply string for any PRIVMSG commands.
	CmdCap      = "CAP"      // IRCv3 Capability negotiation.
	CmdConnect  = "CONNECT"  // Request a new connection to another server immediately.
	CmdDie      = "DIE"      // Shutdown the server.
	CmdError    = "ERROR"    // Report a serious or fatal error to a peer.
	CmdInfo     = "INFO"     // Get information describing a server.
	CmdInvite   = "INVITE"   // Invite a user to a channel.
	CmdIsOn     = "ISON"     // Determine if a nickname is currently on IRC.
	CmdJoin     = "JOIN"     // Join a channel.
	CmdKick     = "KICK"     // Request the forced removal of a user from a channel.
	CmdKill     = "KILL"     // Close a client-server connection by the server which has the actual connection.
	CmdLinks    = "LINKS"    // List all servernames which are known by the server answering the query.
	CmdList     = "LIST"     // List channels and their topics.
	CmdLUsers   = "LUSERS"   // Get statistics about the size of the IRC network.
	CmdMode     = "MODE"     // User mode.
	CmdMOTD     = "MOTD"     // Get the Message of the Day.
	CmdNames    = "NAMES"    // List all visible nicknames.
	CmdNick     = "NICK"     // ":<newnick>" Define a nickname.
	CmdNJoin    = "NJOIN"    // Exchange the list of channel members for each channel between servers.
	CmdNotice   = "NOTICE"   // Send a notice message to specific users or channels.
	CmdOper     = "OPER"     // Obtain operator privileges.
	CmdPart     = "PART"     // Leave a channel.
	CmdPass     = "PASS"     // Set a connection password.
	CmdPing     = "PING"     // Test for the presence of an active client or server.
	CmdPong     = "PONG"     // Reply to a PING message.
	CmdPrivmsg  = "PRIVMSG"  // Send private messages between users, as well as to send messages to channels.
	CmdQuit     = "QUIT"     // Terminate the client session.
	CmdRehash   = "REHASH"   // Force the server to re-read and process its configuration file.
	CmdRestart  = "RESTART"  // Force the server to restart itself.
	CmdServer   = "SERVER"   // Register a new server.
	CmdService  = "SERVICE"  // Register a new service.
	CmdServList = "SERVLIST" // List services currently connected to the network.
	CmdSQuery   = "SQUERY"   //
	CmdSQuit    = "SQUIT"    // Break a local or remote server link.
	CmdStats    = "STATS"    // Get server statistics.
	CmdTagMsg   = "TAGMSG"   // https://ircv3.net/specs/extensions/message-tags.html
	CmdTime     = "TIME"     // Get the local time from the specified server.
	CmdTopic    = "TOPIC"    // Change or view the topic of a channel.
	CmdTrace    = "TRACE"    // Find the route to a server and information about it's peers.
	CmdUser     = "USER"     // Specify the username, hostname and realname of a new user.
	CmdUserHost = "USERHOST" // Get a list of information about upto 5 nicknames.
	CmdUsers    = "USERS"    // Get a list of users logged into the server.
	CmdVersion  = "VERSION"  // Get the version of the server program.
	CmdWAllOps  = "WALLOPS"  // Send a message to all currently connected users who have set the 'w' user mode.
	CmdWho      = "WHO"      // List a set of users.
	CmdWhoIs    = "WHOIS"    // Get information about a specific user.
	CmdWhoWas   = "WHOWAS"   // Get information about a nickname which no longer exists.
)

// irc connection reply codes.
const (
	RplWelcome  = "001" // "Welcome to the Internet Relay Network <nick>!<user>@<host>"
	RplYourHost = "002" // "Your host is <servername>, running version <ver>"
	RplCreated  = "003" // "This server was created <date>"
	RplMyInfo   = "004" // "<servername> <version> <available user modes> <available channel modes>"
	RplISupport = "005" // http://www.irc.org/tech_docs/005.html http://www.irc.org/tech_docs/draft-brocklesby-irc-isupport-03.txt https://www.mirc.com/isupport.html
	RplBounce   = "010" // "Try server <server name>, port <port number>" - https://modern.ircdocs.horse/#rplbounce-010
)

// irc command reply codes.
const (
	RplTraceLink       = "200" // "Link <version & debug level> <destination><next server> V<protocol version> <link uptime in seconds><backstream sendq> <upstream sendq>"
	RplTraceConnecting = "201" // "Try. <class> <server>"
	RplTraceHandshake  = "202" // "H.S. <class> <server>"
	RplTraceUnknown    = "203" // "???? <class> [<client IP address in dot form>]"
	RplTraceOperator   = "204" // "Oper <class> <nick>"
	RplTraceUser       = "205" // "User <class> <nick>"
	RplTraceServer     = "206" // "Serv <class> <int>S <int>C <server><nick!user|*!*>@<host|server> V<protocol version>"
	RplTraceService    = "207" // "Service <class> <name> <type> <activetype>"
	RplTraceNewtype    = "208" // "<newtype> 0 <client name>"
	RplTraceClass      = "209" // "Class <class> <count>"
	RplTraceReconnect  = "210" // Unused.
	RplStatsLinkInfo   = "211" // "<linkname> <sendq> <sent messages> <sentKbytes> <received messages> <received Kbytes> <timeopen>"
	RplStatsCommands   = "212" // "<command> <count> <byte count> <remotecount>"
	RplEndOfStats      = "219" // "<stats letter> :End of STATS report"
	RplUModeIs         = "221" // "<user mode string>"
	RplServList        = "234" // "<name> <server> <mask> <type><hopcount> <info>"
	RplServListEnd     = "235" // "<mask> <type> :End of service listing"
	RplStatsUptime     = "242" // ":Server Up %d days %d:%02d:%02d"
	RplStatsOLine      = "243" // "O <hostmask> * <name>"
	RplLUserClient     = "251" // ":There are <integer> users and <integer> services on<integer> servers"
	RplLUserOp         = "252" // "<integer> :operator(s) online"
	RplLUserUknownL    = "253" // "<integer> :unknown connection(s)"
	RplLUserChannels   = "254" // "<integer> :channels formed"
	RplLUserMe         = "255" // ":I have <integer> clients and <integer> servers"
	RplAdminMe         = "256" // "<server> :Administrative info"
	RplAdminLoc1       = "257" // ":<admin info>"
	RplAdminLoc2       = "258" // ":<admin info>"
	RplAdminEmail      = "259" // ":<admin info>"
	RplTraceLog        = "261" // "File <logfile> <debug level>"
	RplTraceEnd        = "262" // "<server name> <version & debug level> :End of TRACE"
	RplTryAgain        = "263" // "<command> :Please wait a while and try again."
	RplAway            = "301" // "<nick> :<away message>"
	RplUserHost        = "302" // ":*1<reply> *( " " <reply> )"
	RplIsOn            = "303" // ":*1<nick> *( " " <nick> )"
	RplUnAway          = "305" // ":You are no longer marked as being away"
	RplNowAway         = "306" // ":You have been marked as being away"
	RplWhoIsUser       = "311" // "<nick> <user> <host> * :<real name>"
	RplWhoIsServer     = "312" // "<nick> <server> :<server info>"
	RplWhoIsOperator   = "313" // "<nick> :is an IRC operator"
	RplWhoWasUser      = "314" // "<nick> <user> <host> * :<real name>"
	RplEndOfWho        = "315" // "<name> :End of WHO list"
	RplWhoIsIdle       = "317" // "<nick> <integer> :seconds idle"
	RplEndOfWhoIs      = "318" // "<nick> :End of WHOIS list"
	RplWhoIsChannels   = "319" // "<nick> :*( ( "@" / "+" ) <channel>" " )"
	RplListStart       = "321" // Obsolete.
	RplList            = "322" // "<channel> <# visible> :<topic>"
	RplListEnd         = "323" // ":End of LIST"
	RplChannelModeIs   = "324" // "<channel> <mode> <mode params>"
	RplUniqOpIs        = "325" // "<channel> <nickname>"
	RplNoTopic         = "331" // "<channel> :No topic is set"
	RplTopic           = "332" // "<channel> :<topic>"
	RplWhoisBot        = "335" // "<nick> <target> :<message>"
	RplInviting        = "341" // "<channel> <nick>"
	RplSummoning       = "342" // "<user> :Summoning user to IRC"
	RplInviteList      = "346" // "<channel> <invitemask>"
	RplEndOfInviteList = "347" // "<channel> :End of channel invite list"
	RplExceptList      = "348" // "<channel> <exceptionmask>"
	RplEndOfExceptList = "349" // "<channel> :End of channel exception list"
	RplVersion         = "351" // "<version>.<debuglevel> <server>:<comments>"
	RplWhoReply        = "352" // "<channel> <user> <host> <server><nick> ( "H" / "G" > ["*"] [ ("@" / "+" ) ] :<hopcount> <real name>"
	RplNamReply        = "353" // "( "=" / "*" / "@" ) <channel>:[ "@" / "+" ] <nick> *( " " ["@" / "+" ] <nick> )"
	RplLinks           = "364" // "<mask> <server> :<hopcount> <serverinfo>"
	RplEndOfLinks      = "365" // "<mask> :End of LINKS list"
	RplEndOfNames      = "366" // "<channel> :End of NAMES list"
	RplBanList         = "367" // "<channel> <banmask>"
	RplEndOfBanList    = "368" // "<channel> :End of channel ban list"
	RplEndOfWhoWas     = "369" // "<nick> :End of WHOWAS"
	RplInfo            = "371" // ":<string>"
	RplMOTD            = "372" // ":- <text>"
	RplEndOfInfo       = "374" // ":End of INFO list"
	RplMOTDStart       = "375" // ":- <server> Message of the day - "
	RplEndOfMOTD       = "376" // ":End of MOTD command"
	RplYoureOper       = "381" // ":You are now an IRC operator"
	RplRehashing       = "382" // "<config file> :Rehashing"
	RplYoureService    = "383" // "You are service <servicename>"
	RplTime            = "391" // "<server> :<string showing server's local time>"
	RplUsersStart      = "392" // ":UserID Terminal Host"
	RplUsers           = "393" // ":<username> <ttyline> <hostname>"
	RplEndOfUsers      = "394" // ":End of users"
	RplNoUsers         = "395" // ":Nobody logged in"
	RplHostHidden      = "396" // "fubarbot <host> :is now your displayed host" Reply to a user when user mode +x (host masking) was set successfully https://www.alien.net.au/irc/irc2numerics.html
)

// irc error reply codes.
const (
	RplErrNoSuchNick        = "401" // "<nickname> :No such nick/channel"
	RplErrNoSuchServer      = "402" // "<server name> :No such server"
	RplErrNoSuchChannel     = "403" // "<channel name> :No such channel"
	RplErrCannotSendToChan  = "404" // "<channel name> :Cannot send to channel"
	RplErrTooManyChannels   = "405" // "<channel name> :You have joined too many channels"
	RplErrWasNoSuchNick     = "406" // "<nickname> :There was no such nickname"
	RplErrTooManyTargets    = "407" // "<target> :<error code> recipients. <abortmessage>"
	RplErrNoSuchService     = "408" // "<service name> :No such service"
	RplErrNoOrigin          = "409" // ":No origin specified"
	RplErrInvalidCapCmd     = "410"
	RplErrNoRecipient       = "411" // ":No recipient given (<command>)"
	RplErrNoTextToSend      = "412" // ":No text to send"
	RplErrNoToplevel        = "413" // "<mask> :No toplevel domain specified"
	RplErrWildToplevel      = "414" // "<mask> :Wildcard in toplevel domain"
	RplErrBadMask           = "415" // "<mask> :Bad Server/host mask"
	RplErrUnknownCommand    = "421" // "<command> :Unknown command"
	RplErrNoMOTD            = "422" // ":MOTD File is missing"
	RplErrNoAdminInfo       = "423" // "<server> :No administrative info available"
	RplErrFileError         = "424" // ":File error doing <file op> on <file>"
	RplErrNoNicknameGiven   = "431" // ":No nickname given"
	RplErrErroneousNickname = "432" // "<client> <nick> :Erroneus nickname"
	RplErrNicknameInUse     = "433" // "<client> <nick> :Nickname is already in use"
	RplErrNickCollision     = "436" // "<nick> :Nickname collision KILL from<user>@<host>"
	RplErrUnavailResource   = "437" // "<nick/channel> :Nick/channel is temporarily unavailable"
	RplErrUserNotInChannel  = "441" // "<nick> <channel> :They aren't on that channel"
	RplErrNotOnChannel      = "442" // "<channel> :You're not on that channel"
	RplErrUserOnChannel     = "443" // "<user> <channel> :is already on channel"
	RplErrNoLogin           = "444" // "<user> :User not logged in"
	RplErrSummonDisabled    = "445" // ":SUMMON has been disabled"
	RplErrUsersDisabled     = "446" // ":USERS has been disabled"
	RplErrNotRegistered     = "451" // ":You have not registered"
	RplErrNeedMoreParams    = "461" // "<command> :Not enough parameters"
	RplErrAlreadyRegistered = "462" // ":Unauthorized command (already registered)"
	RplErrNoPermForHost     = "463" // ":Your host isn't among the privileged"
	RplErrPasswdMismatch    = "464" // ":Password incorrect"
	RplErrYoureBannedCreep  = "465" // ":You are banned from this server"
	RplErrYouWillBeBanned   = "466" //
	RplErrKeySet            = "467" // "<channel> :Channel key already set"
	RplErrChannelIsFull     = "471" // "<channel> :Cannot join channel (+l)"
	RplErrUnknownMode       = "472" // "<char> :is unknown mode char to me for <channel>"
	RplErrInviteOnlyChan    = "473" // "<channel> :Cannot join channel (+i)"
	RplErrBannedFromChan    = "474" // "<channel> :Cannot join channel (+b)"
	RplErrBadChannelKey     = "475" // "<channel> :Cannot join channel (+k)"
	RplErrBadChanMask       = "476" // "<channel> :Bad Channel Mask"
	RplErrNoChanModes       = "477" // "<channel> :Channel doesn't support modes"
	RplErrBanListFull       = "478" // "<channel> <char> :Channel list is full"
	RplErrNoPrivileges      = "481" // ":Permission Denied- You're not an IRC operator"
	RplErrChanOPrivsNeeded  = "482" // "<channel> :You're not channel operator"
	RplErrCantKillServer    = "483" // ":You can't kill a server!"
	RplErrRestricted        = "484" // ":Your connection is restricted!"
	RplErrUniqOPrivsNeeded  = "485" // ":You're not the original channel operator"
	RplErrNoOperHost        = "491" // ":No O-lines for your host"
	RplErrUModeUnknownFlag  = "501" // ":Unknown MODE flag"
	RplErrUsersDontMatch    = "502" // ":Cannot change mode for other users"
)

// Client-to-Client Protocol command constants. These commands are NOT sent by the server; they are instead generated
// internally as replacements for CTCP-formatted PRIVMSG and NOTICE messages.
//
// For convenience, these constants are defined for well-known CTCP messages. To create handlers that
// match arbitrary CTCP commands and replies, see NewCTCPCmd and NewCTCPReplyCmd.
const (
	CTCPAction = "_CTCP_QUERY_ACTION"

	CTCPVersionQuery    = "_CTCP_QUERY_VERSION"
	CTCPVersionReply    = "_CTCP_REPLY_VERSION"
	CTCPClientInfoQuery = "_CTCP_QUERY_CLIENTINFO"
	CTCPClientInfoReply = "_CTCP_REPLY_CLIENTINFO"

	CTCPPingQuery = "_CTCP_QUERY_PING"
	CTCPPingReply = "_CTCP_REPLY_PING"
	CTCPTimeQuery = "_CTCP_QUERY_TIME"
	CTCPTimeReply = "_CTCP_REPLY_TIME"
)
