from fluffy import app

# import views so the decorators run
import fluffy.views  # noreorder # noqa


if __name__ == '__main__':
    app.run(debug=True)
