class AppView extends KDView

  constructor:->

    super

    app = @getData()

    @followButton = new KDToggleButton
      style           : "kdwhitebtn"
      dataPath        : "followee"
      defaultState    : if app.followee then "Unfollow" else "Follow"
      states          : [
        "Follow", (callback)->
          app.follow (err)->
            callback? err
        "Unfollow", (callback)->
          app.unfollow (err)->
            callback? err
      ]
    , app

    @likeButton = new KDToggleButton
      style           : "kdwhitebtn"
      states          : [
        "Like", (callback)->
          app.like (err)->
            callback? err
        "Unlike", (callback)->
          app.like (err)->
            callback? err
      ]
    , app

    appsController = @getSingleton("kodingAppsController")

    if KD.checkFlag 'super-admin'
      @approveButton = new KDToggleButton
        style           : "kdwhitebtn"
        dataPath        : "approved"
        defaultState    : if app.approved then "Disapprove" else "Approve"
        states          : [
          "Approve", (callback)->
            appsController.approveApp app, (err)=>
              if not err
                app.approve yes, (err)->
                  if err then warn err
                  callback? err
              else
                callback? err
          "Disapprove", (callback)->
            app.approve no, (err)->
              callback? err
        ]
      , app

      @removeButton = new KDButtonView
        title    : "Delete"
        style    : "kdwhitebtn"
        callback : =>
          modal = new KDModalView
            title          : "Delete App"
            content        : "<div class='modalformline'>Are you sure you want to delete this application?</div>"
            height         : "auto"
            overlay        : yes
            buttons        :
              Delete       :
                style      : "modal-clean-red"
                loader     :
                  color    : "#ffffff"
                  diameter : 16
                callback   : =>
                  app.delete (err)=>
                    modal.buttons.Delete.hideLoader()
                    modal.destroy()
                    if not err
                      @emit 'AppDeleted', app
                      appManager.openApplication "Apps", yes, (instance)=>
                        @utils.wait 100, instance.feedController.changeActiveSort "meta.modifiedAt"
                        callback?()
                    else
                      new KDNotificationView
                        type     : "mini"
                        cssClass : "error editor"
                        title    : "Error, please try again later!"
                      warn err

    else
      @approveButton = new KDView
      @removeButton  = new KDView

    app.checkIfLikedBefore (err, likedBefore)=>
      if likedBefore
        @likeButton.setState "Unlike"
      else
        @likeButton.setState "Like"


    if app.versions?.length > 1
      menu = {}

      for version,i in app.versions
        menu["Install version #{version}"] =
          callback : (item)=>
            appsController.installApp app, app.versions[item.data.id], (err)=>
              if err then warn err

      @installButton = new KDButtonViewWithMenu
        title     : "Install Now"
        style     : "cupid-green"
        loader    :
          top     : 0
          diameter: 30
          color   : "#ffffff"
        delegate      : @
        menu          : menu
        callback      : ->
          appsController.installApp app, 'latest', (err)=>
            @hideLoader()

    else
      @installButton = new KDButtonView
        title     : "Install Now"
        style     : "cupid-green"
        loader    :
          top     : 0
          diameter: 30
          color   : "#ffffff"
        callback  : ->
          appsController.installApp app, 'latest', (err)=>
            @hideLoader()

    # # @forkButton = new KDButtonView
    #   title     : "Fork"
    #   style     : "clean-gray"
    #   disabled  : !app.manifest.repo?
    #   loader    :
    #     top     : 0
    #     diameter: 30
    #   callback  : ->
    #     appsController.forkApp app, (err)=>
    #       @hideLoader()

    # unless app.manifest.repo
    #   @forkButton.setTooltip
    #     title   : "No repository specified for the app!"
    #     gravity : "w"

  viewAppended:->
    @setTemplate @pistachio()
    @template.update()

  putThumb:(manifest)->

    {icns, name, version, authorNick} = manifest
    src = if icns and (icns['256'] or icns['512'] or icns['128'] or icns['160'] or icns['64'])
      "#{KD.appsUri}/#{authorNick}/#{name}/#{version}/#{if icns then icns['256'] or icns['512'] or icns['128'] or icns['160'] or icns['64']}"
    else
      "#{KD.apiUri + '/images/default.app.thumb.png'}"

    return "<img src='#{src}'/>"

  pistachio:->
    """
    <div class="profileleft">
      <span>
        <a class='profile-avatar' href='#'>{{ @putThumb #(manifest)}}</a>
      </span>
    </div>
    <section class="right-overflow">
      <h3 class='profilename'>{{#(title)}}<cite>by {{#(manifest.author)}}</cite></h3>
      <div class="installerbar clearfix">
        {{> @installButton}}
        <div class="versionstats updateddate">Version {{ #(manifest.version) or "---" }}<p>Updated: ---</p></div>
        <div class="versionscorecard">
          <div class="versionstats">{{#(counts.installed) or 0}}<p>INSTALLS</p></div>
          <div class="versionstats">{{#(meta.likes) or 0}}<p>Likes</p></div>
          <div class="versionstats">{{#(counts.followers) or 0}}<p>Followers</p></div>
        </div>
        <div class="appfollowlike">
          {{> @followButton}}
          {{> @likeButton}}
        </div>
        <div class="appfollowlike">
          {{> @approveButton}}
          {{> @removeButton}}
        </div>
      </div>
    </section>
    """

class AppViewOld extends KDView

  constructor:->

    super

    app = @getData()

    @followButton = new KDToggleButton
      style           : "kdwhitebtn"
      dataPath        : "followee"
      states          : [
        "Follow", (callback)->
          callback? null
        "Unfollow", (callback)->
          callback? null
      ]
    , app

    @likeButton = new KDToggleButton
      style           : "kdwhitebtn"
      dataPath        : "followee"
      states          : [
        "Like", (callback)->
          callback? null
        "Unlike", (callback)->
          callback? null
      ]
    , app

    @installButton = new KDButtonView
      title     : "Install Now"
      style     : "cupid-green"
      callback  : =>
        modalOptions = @sanitizeInstallModalOptions()
        modal = new KDModalViewWithForms
          title         : "Application configuration"
          width         : 500
          height        : "auto"
          overlay       : yes
          tabs          :
            navigable   : yes
            callback    : (formOutput)=>
              @createBashScript formOutput,app
              modal.destroy()
            forms       : modalOptions


  createBashScript:(formOutput,app)->
    bashScript = Encoder.htmlDecode app.attachments[0].content
    for key,value of formOutput
      bashScript = bashScript.replace "$#{key}","\"#{value}\""

    log bashScript

  sanitizeInstallModalOptions:->
    app     = @getData()
    reqs    = app.attachments[1].content
    modalOptions = JSON.parse Encoder.htmlDecode reqs
    count   = 0
    for tabName,options of modalOptions
      unless Object.keys(modalOptions).length - 1 is count
        options.buttons =
          Next    :
            title : "Next"
            style : "modal-clean-gray"
            type  : "submit"
      else
        options.buttons =
          Install :
            title : "INSTALL THE APP!"
            style : "cupid-green"
            type  : "submit"
      count++
    modalOptions

  viewAppended:->
    @setTemplate @pistachio()
    @template.update()

  putThumb:(thumbnails = [])->
    if thumbnails.length > 0
      thumb = "<img src='/images/uploads/#{thumbnails[0].appThumb}'/>"
    else
      ""

  pistachio:->
    """
    <div class="profileleft">
      <span>
        <a class='profile-avatar' href='#'>{{ @putThumb #(thumbnails)}}</a>
      </span>
    </div>
    <section class="right-overflow">
      {h3.profilename{#(title)}}
      <div class="installerbar clearfix">
        {{> @installButton}}
        <div class="versionstats updateddate">Version ---<p>Updated: ---</p></div>
        <div class="versionscorecard">
          <div class="versionstats">{{#(counts.installed)}}<p>INSTALLS</p></div>
          <div class="versionstats">0{{#(counts.likes)}}<p>Likes</p></div>
          <div class="versionstats">{{#(counts.followers)}}<p>Followers</p></div>
        </div>
        <div class="appfollowlike">
          {{> @followButton}}
          {{> @likeButton}}
        </div>
      </div>
    </section>
    """
