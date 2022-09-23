import {Component} from '@angular/core';
import {Feature} from "./services/feature";
import {FeatureService} from "./services/feature.service";

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
})
export class AppComponent {
  features: Feature[] = [];

  constructor(
    private featureService: FeatureService,
  ) {
  }

  ngOnInit(): void {
    this.getFeatures();
  }

  getFeatures(): void {
    this.featureService.getFeatures()
      .subscribe(features => this.features = features);
  }
}
