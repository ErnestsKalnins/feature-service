import {Component, OnInit} from '@angular/core';
import {Location} from "@angular/common";
import {Feature} from "../services/feature";
import {FeatureService} from "../services/feature.service";
import {Router} from "@angular/router";

@Component({
  selector: 'app-new-feature',
  templateUrl: './new-feature.component.html',
})
export class NewFeatureComponent implements OnInit {
  feature: Feature = {
    id: null,
    technicalName: '',
    displayName: null,
    description: null,
    expiresOn: null,
    inverted: false,
  };

  constructor(
    private featureService: FeatureService,
    private location: Location,
    private router: Router
  ) {
  }

  ngOnInit(): void {
  }

  goBack(): void {
    this.location.back();
  }

  invert(): void {
    this.feature.inverted = !this.feature.inverted;
  }

  saveFeature(): void {
    const that = this;
    this.featureService.saveFeature(this.feature)
      .subscribe({
        complete() {
          that.router.navigate(['/features']);
        }
      });
  }
}
