class Inbox12345 extends AppController

  {race} = Bongo

  constructor:(options, data)->
    view = new (KD.getPageClass 'Inbox') cssClass : "inbox-application"
    options = $.extend {view},options
    super options,data
    @selection = {}

  bringToFront:()->
    @propagateEvent (KDEventType : 'ApplicationWantsToBeShown', globalEvent : yes),
      options :
        name : 'Inbox'
      data : @mainView

  initAndBringToFront:(options,callback)->
    initApplication options, ->
      @bringToFront()
      callback()

  initApplication:(options, callback)->
    callback()

  fetchMessages:(options, callback)->
    #KD.whoami().fetchMail? options, callback

  fetchAutoCompleteForToField:(inputValue,blacklist,callback)->
    KD.remote.api.JAccount.byRelevance inputValue,{blacklist},(err,accounts)->
      callback accounts

  loadView:(mainView)->
    mainView.createCommons()
    mainView.createTabs()

    {currentDelegate} = KD.getSingleton('mainController').getVisitor()

    mainView.registerListener
      KDEventTypes : "ToFieldHasNewInput"
      listener     : @
      callback     : (pubInst, data)->
        return if data.disabledForBeta
        {type,action} = data
        mainView.showTab type
        if action is "change-tab"
          mainView.showTab data.type
        else
          mainView.sort data.type

    mainView.registerListener
      KDEventTypes      : 'NotificationIsSelected'
      listener          : @
      callback          :(pubInst, {notification, event, location})=>
        # nothing yet, coming soon

    {newMessageBar} = mainView

    mainView.registerListener
      KDEventTypes: 'MessageIsSelected'
      listener    : @
      callback    :(pubInst,{item, event})=>
        data = item.getData()
        data.mark 'read', (err)->
          item.unsetClass 'unread' unless err
        # unless event.shiftKey
        #   @deselectMessages()
        @deselectMessages()
        if item.paneView?
          {paneView} = item
          mainView.inboxMessagesContainer.showPane item.paneView
        else
          paneView = new KDTabPaneView
            name: data.subject
            hiddenHandle: yes
          mainView.inboxMessagesContainer.addPane paneView
          detail = new InboxMessageDetail cssClass : "message-detail", data

          detail.registerListener
            KDEventTypes: 'viewAppended'
            listener: @
            callback: =>
              data.restComments 0, (err, comments)-> # log arguments, data

          paneView.addSubView detail
          paneView.detail = detail
          item.paneView = paneView
        newMessageBar.enableMessageActionButtons()
        @selectMessage data, item, paneView

    newMessageBar.registerListener
      KDEventTypes  : "AutoCompleteNeedsMemberData"
      listener      : @
      callback      : (pubInst,event)=>
        {callback,inputValue,blacklist} = event
        @fetchAutoCompleteForToField inputValue,blacklist,callback

    newMessageBar.registerListener
      KDEventTypes  : 'MessageShouldBeSent'
      listener      : @
      callback      : (pubInst,{formOutput,callback})->
        @prepareMessage formOutput, callback, newMessageBar

    newMessageBar.registerListener
      KDEventTypes: 'MessageShouldBeDisowned'
      listener    : @
      callback    : do=>
        if not @selection
          newMessageBar.disableMessageActionButtons()
          modal.destroy()
          return
        disownAll = (items, callback)->
          disownItem = race (i, item, fin)->
            item.data.disown (err)->
              if err
                fin err
              else
                fin()
          , callback
          disownItem item for own id, item of items
        (pubInst, modal) =>
          disownAll @selection, =>
            for own id, {item, paneView} of @selection
              item.destroy()
              paneView.destroy()
              @deselectMessages()
            modal.destroy()
            newMessageBar.disableMessageActionButtons()

    newMessageBar.registerListener
      KDEventTypes: 'MessageShouldBeMarkedAsUnread'
      listener    : @
      callback    : =>
        for own id, {item, data} of @selection
          data.unmark 'read', (err)=>
            log err if err
            item.setClass 'unread' unless err
            item.paneView?.hide()
            newMessageBar.disableMessageActionButtons()

  goToNotifications:(notification)->
    @getView().showTab "notifications"
    @mainView.propagateEvent KDEventType : 'NotificationIsSelected', {item: notification, event} if notification?

  goToMessages:(message)->
    @getView().showTab "messages"
    @mainView.emit 'MessageSelectedFromOutside', message
    
  selectMessage:(data, item, paneView)->
    @selection[data.getId()] = {
      data
      item
      paneView
    }

  deselectMessages:->
    @selection = {}

  sendMessage:(messageDetails, callback)->
    # log "I just send a new message: ", messageDetails
    KD.remote.api.JPrivateMessage.create messageDetails, callback

  prepareMessage:(formOutput, callback, newMessageBar)=>
    {body, subject, recipients} = formOutput

    to = recipients.join ' '

    @sendMessage {to, body, subject}, (err, message)=>
      new KDNotificationView
        title     : if err then "There was an error sending your message - try again" else "Message Sent!"
        duration  : 1000
      message.mark 'read'
      newMessageBar.emit 'RefreshButtonClicked'
      callback? err, message
